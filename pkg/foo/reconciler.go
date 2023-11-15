package foo

import (
	"context"

	"github.com/grafana/grafana-app-sdk/operator"

	foo "github.com/radiohead/grafana-sdk-example/pkg/generated/resource/foo/v1alpha1"
	"github.com/radiohead/grafana-sdk-example/pkg/instrument"
)

// Reconciler reconciles foo.Object events and resyncs,
// by updating the local in-memory cache.
type Reconciler struct {
	cache *Cache
}

// NewReconciler returns a new Reconciler which uses provided Cache.
func NewReconciler(cache *Cache) *Reconciler {
	return &Reconciler{
		cache: cache,
	}
}

// Reconcile handles the request by syncing the object to the cache.
func (r *Reconciler) Reconcile(
	ctx context.Context, req operator.TypedReconcileRequest[*foo.Object],
) (operator.ReconcileResult, error) {
	ctx, logger, span := instrument.StartSpan(ctx, "foo.Reconciler#Reconcile")
	defer span.End()

	switch action := req.Action; action {
	case operator.ReconcileActionCreated:
		r.cache.Add(ctx, *req.Object)
		logger.Debug("processed create action")
	case operator.ReconcileActionUpdated:
		r.cache.Add(ctx, *req.Object)
		logger.Debug("processed update action")
	case operator.ReconcileActionDeleted:
		r.cache.Delete(ctx, *req.Object)
		logger.Debug("processed delete action")
	case operator.ReconcileActionResynced:
		r.cache.Add(ctx, *req.Object)
		logger.Debug("processed resync action")
	default:
		logger.Info(
			"unknown reconcile action received",
			"action", action,
		)
	}

	return operator.ReconcileResult{}, nil
}
