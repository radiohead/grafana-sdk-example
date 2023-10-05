package instrument

import (
	"context"

	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/grafana/grafana-app-sdk/operator"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan starts a new tracing span and logging context with the given name from the provided context.
func StartSpan(ctx context.Context, spanName string, logArgs ...any) (context.Context, logging.Logger, trace.Span) {
	ctx, span := operator.GetTracer().Start(ctx, spanName)

	ln := len(logArgs) + 2
	args := make([]any, ln)
	args[0] = "context"
	args[1] = spanName
	copy(args[2:ln], logArgs)

	logger := logging.FromContext(ctx).With(args...)

	return ctx, logger, span
}
