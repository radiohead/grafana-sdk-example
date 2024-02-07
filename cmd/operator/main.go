package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/grafana/grafana-app-sdk/k8s"
	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/grafana/grafana-app-sdk/metrics"
	"github.com/grafana/grafana-app-sdk/operator"
	"github.com/grafana/grafana-app-sdk/plugin/kubeconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/radiohead/grafana-sdk-example/pkg/foo"
	fookind "github.com/radiohead/grafana-sdk-example/pkg/generated/resource/foo/v1alpha1"
)

const (
	leaseIdentifier = "foo-operator"
	leaseName       = "foo-operator-lease"
)

func main() {
	fmt.Fprintf(os.Stderr, "operator starting")

	if err := run(); err != nil {
		// TODO: maybe change default SDK logger to write to STDOUT / STDERR.
		logger := logging.DefaultLogger
		if _, ok := logger.(*logging.NoOpLogger); ok {
			fmt.Fprintf(
				os.Stderr,
				"error running operator err=%s",
				err.Error(),
			)
		} else {
			logger.Error(
				"error running operator",
				"err", err,
			)
		}

		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "operator stopped")

	os.Exit(0)
}

func run() error { // nolint: funlen
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}

	logging.DefaultLogger = logging.NewSLogLogger(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	if err := SetTraceProvider(cfg.OTelConfig); err != nil {
		return err
	}

	kubeConfig, err := LoadKubeconfig(cfg.KubeconfigVal, cfg.KubeconfigPath)
	if err != nil {
		return err
	}

	runner, err := newRunner(cfg, kubeConfig)
	if err != nil {
		return err
	}

	// Set up metrics exporter
	exporter := metrics.NewExporter(metrics.ExporterConfig{})
	if err := exporter.RegisterCollectors(runner.PrometheusCollectors()...); err != nil {
		return err
	}
	runner.AddController(exporter)

	// We have to create a vanilla k8s client for leader election.
	client := clientset.NewForConfigOrDie(&kubeConfig.RestConfig)

	// TODO: re-write to return an error instead of panicking.
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Name:      leaseName,
				Namespace: kubeConfig.Namespace,
			},
			Client: client.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: leaseIdentifier,
			},
		},
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				defer cancel()

				logging.FromContext(ctx).Info(
					"instance started leading, starting the operator...",
				)

				if err := runner.Run(ctx.Done()); err != nil {
					logging.FromContext(ctx).Error(
						"error running operator",
						"error", err,
					)
				}

				logging.FromContext(ctx).Info(
					"operator finished running, will exit...",
				)
			},
			OnStoppedLeading: func() {
				defer cancel()

				logging.FromContext(ctx).Info(
					"instance stopped leading, will exit...",
				)
			},
		},
	})

	return nil
}

func newRunner(
	operatorCfg *Config,
	kubeConfig kubeconfig.NamespacedConfig,
) (*operator.Operator, error) {
	cache := foo.NewCache(1024)

	clientGenerator := k8s.NewClientRegistry(kubeConfig.RestConfig, k8s.ClientConfig{})
	fooClient, err := clientGenerator.ClientFor(fookind.Schema())
	if err != nil {
		return nil, err
	}

	rec, err := operator.NewOpinionatedReconciler(fooClient, "foo-operator")
	if err != nil {
		return nil, err
	}

	rec.Wrap(&operator.TypedReconciler[*fookind.Object]{
		ReconcileFunc: foo.NewReconciler(cache).Reconcile,
	})

	controller := operator.NewInformerController(operator.DefaultInformerControllerConfig())
	if err := controller.AddReconciler(rec, fookind.Schema().Kind()); err != nil {
		return nil, err
	}

	fooInformer, err := operator.NewKubernetesBasedInformer(fookind.Schema(), fooClient, kubeConfig.Namespace)
	if err != nil {
		return nil, err
	}

	if err = controller.AddInformer(fooInformer, fookind.Schema().Kind()); err != nil {
		return nil, err
	}

	srv, err := k8s.NewWebhookServer(k8s.WebhookServerConfig{
		Port: operatorCfg.WebhookServer.Port,
		TLSConfig: k8s.TLSConfig{
			CertPath: operatorCfg.WebhookServer.TLSCertPath,
			KeyPath:  operatorCfg.WebhookServer.TLSKeyPath,
		},
	})
	if err != nil {
		return nil, err
	}
	srv.AddValidatingAdmissionController(foo.NewValidator(cache), fookind.Schema())

	res := operator.New()
	res.AddController(controller)
	res.AddController(srv)

	return res, nil
}
