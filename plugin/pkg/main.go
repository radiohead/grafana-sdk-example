package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/grafana/grafana-app-sdk/k8s"
	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/grafana/grafana-app-sdk/metrics"
	sdkPlugin "github.com/grafana/grafana-app-sdk/plugin"
	"github.com/grafana/grafana-app-sdk/plugin/kubeconfig"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/app"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"

	"github.com/radiohead/grafana-sdk-example/pkg/generated/resource/foo"
)

const (
	pluginID = "grafana-sdk-example"
)

var (
	enabledResources = []resource.Schema{
		foo.Schema(),
	}
)

func main() {
	// Set the app-sdk logger to use the plugin-sdk logger
	logger := sdkPlugin.NewLogger(log.DefaultLogger.With("pluginID", pluginID))
	logger.Info("starting plugin", "pluginID", pluginID)
	logging.DefaultLogger = logger

	// TODO: we need to revise schema group structure.
	// 1. AddSchema should support versions (intead of hard-coding single version for all resources).
	// 2. We should use codegen for generating schema groups (once we move away from Thema).
	schemaGroup := resource.NewSimpleSchemaGroup(foo.Schema().Group(), foo.Schema().Version())
	for _, r := range enabledResources {
		schemaGroup.AddSchema(
			r.ZeroValue(),
			resource.WithKind(r.Kind()),
			resource.WithPlural(r.Plural()),
			resource.WithScope(resource.NamespacedScope),
		)
	}

	// app.Manage handles the app plugin lifecycle
	if err := app.Manage(pluginID, newInstanceFactory(schemaGroup, logger), app.ManageOpts{
		TracingOpts: tracing.Opts{
			CustomAttributes: []attribute.KeyValue{
				attribute.String("plugin.id", pluginID),
			},
		},
	}); err != nil {
		logger.Error("failed to initialize instance", "err", err)
		os.Exit(1)
	}

	logger.Info("plugin exited normally", "pluginID", pluginID)
	os.Exit(0)
}

// newInstanceFactory returns an app.InstanceFactoryFunc to be used with app.Manage
func newInstanceFactory(schemaGroup resource.SchemaGroup, logger logging.Logger) app.InstanceFactoryFunc {
	return func(ctx context.Context, settings backend.AppInstanceSettings) (instancemgmt.Instance, error) {
		// Load the kubernetes config from the AppInstanceSettings
		var kcfg kubeconfig.NamespacedConfig
		if err := kubeconfig.NewLoader().LoadFromSettings(settings, &kcfg); err != nil {
			logger.Error("failed to load kubernetes config from settings", "err", err)
			return nil, err
		}

		// Create our client generator, using kubernetes as a store.
		clientGenerator := k8s.NewClientRegistry(kcfg.RestConfig, k8s.ClientConfig{
			MetricsConfig: metrics.Config{
				Namespace: strings.ReplaceAll(pluginID, "-", "_"),
			},
		})
		// TODO: connect to metrics protocol
		prometheus.MustRegister(clientGenerator.PrometheusCollectors()...)

		// Create the plugin, which allows for CallResource requests to it as an instancemgmt.Instance
		p, err := NewPlugin(kcfg.Namespace, clientGenerator, schemaGroup)
		if err != nil {
			logger.Error("failed to create plugin instance", "err", err)
			return nil, fmt.Errorf("failed to create plugin instance: %w", err)
		}

		logger.Info("plugin instance provisioned successfully")
		return p, nil
	}
}
