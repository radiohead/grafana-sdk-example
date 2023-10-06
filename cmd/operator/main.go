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
	foov1 "github.com/radiohead/grafana-sdk-example/pkg/generated/resource/foo"
)

const (
	leaseIdentifier = "foo-operator"
	leaseName       = "foo-operator-lease"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Load the config from the environment
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing config: %s", err.Error())
		os.Exit(1)
	}

	// Set up telemetry
	// Logging
	logging.DefaultLogger = logging.NewSLogLogger(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	logger := logging.FromContext(ctx)

	// Tracing
	if err := SetTraceProvider(cfg.OTelConfig); err != nil {
		logger.Error(
			"error configuring tracing",
			"error", err,
		)
		os.Exit(1)
	}

	logger.Info("starting operator...")

	// Load the kube config
	kubeConfig, err := LoadInClusterConfig()
	if err != nil {
		logger.Error(
			"error loading kubeconfig",
			"error", err,
		)
		os.Exit(1)
	}

	runner, err := newRunner(cfg, kubeConfig)
	if err != nil {
		logger.Error(
			"error initialising the runner",
			"error", err,
		)
		os.Exit(1)
	}

	// Set up metrics exporter
	exporter := metrics.NewExporter(metrics.ExporterConfig{})
	exporter.RegisterCollectors(runner.PrometheusCollectors()...)
	runner.AddController(exporter)

	logger.Info("starting leader election loop")

	// We have to create a vanilla k8s client for leader election.
	client := clientset.NewForConfigOrDie(&kubeConfig.RestConfig)

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

	logger.Info("operator exit complete")

	os.Exit(0)
}

func newRunner(
	operatorCfg *Config,
	kubeConfig kubeconfig.NamespacedConfig,
) (*operator.Operator, error) {
	cache := foo.NewCache(1024)

	clientGenerator := k8s.NewClientRegistry(kubeConfig.RestConfig, k8s.ClientConfig{})
	fooClient, err := clientGenerator.ClientFor(foov1.Schema())
	if err != nil {
		return nil, err
	}

	rec, err := operator.NewOpinionatedReconciler(fooClient, "foo-operator")
	if err != nil {
		return nil, err
	}

	rec.Wrap(&operator.TypedReconciler[*foov1.Object]{
		ReconcileFunc: foo.NewReconciler(cache).Reconcile,
	})

	controller := operator.NewInformerController(operator.DefaultInformerControllerConfig())
	if err := controller.AddReconciler(rec, foov1.Schema().Kind()); err != nil {
		return nil, err
	}

	fooInformer, err := operator.NewKubernetesBasedInformer(foov1.Schema(), fooClient, kubeConfig.Namespace)
	if err != nil {
		return nil, err
	}

	if err = controller.AddInformer(fooInformer, foov1.Schema().Kind()); err != nil {
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
	srv.AddValidatingAdmissionController(foo.NewValidator(cache), foov1.Schema())

	res := operator.New()
	res.AddController(controller)
	res.AddController(srv)

	return res, nil
}
