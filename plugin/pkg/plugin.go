package main

import (
	"context"

	"go.opentelemetry.io/otel"

	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/grafana/grafana-app-sdk/metrics"
	"github.com/grafana/grafana-app-sdk/plugin/router"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/prometheus/client_golang/prometheus"
)

// Plugin is the backend plugin
type Plugin struct {
	router *router.ResourceGroupRouter
}

// New returns a new instance of a Plugin.
// TODO: support multiple schema groups.
func NewPlugin(namespace string, client resource.ClientGenerator, group resource.SchemaGroup) (*Plugin, error) {
	// TODO: instrument the router and the store.
	rgRouter, err := router.NewResourceGroupRouter(group, namespace, client)
	if err != nil {
		return nil, err
	}

	rgRouter.Use(
		router.NewMetricsMiddleware(metrics.DefaultConfig(namespace), prometheus.DefaultRegisterer),
		router.NewTracingMiddleware(otel.GetTracerProvider().Tracer("tracing-middleware")),
		router.NewLoggingMiddleware(logging.DefaultLogger),
	)

	rgRouter.NotFoundHandler = func(
		ctx context.Context,
		req *backend.CallResourceRequest,
		snd backend.CallResourceResponseSender,
	) {
		logging.DefaultLogger.Debug(
			"route not found",
			"path", req.Path,
			"method", req.Method,
			"url", req.URL,
		)

		if err := snd.Send(&backend.CallResourceResponse{
			Status: 404,
		}); err != nil {
			logging.DefaultLogger.Error("error sending response", "err", err)
		}
	}

	return &Plugin{
		router: rgRouter,
	}, nil
}

// Start has the plugin's router start listening over gRPC, and blocks until an unrecoverable error occurs
func (p *Plugin) Start() error {
	return p.router.ListenAndServe()
}

// CallResource allows Plugin to implement grafana-plugin-sdk-go/backend/instancemgmt.Instance for an App plugin,
// Which allows it to be used with grafana-plugin-sdk-go/backend/app.Manage.
// CallResource downstreams all CallResource requests to the router's handler
func (p *Plugin) CallResource(
	ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender,
) error {
	return p.router.CallResource(ctx, req, sender)
}
