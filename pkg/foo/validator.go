package foo

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/radiohead/grafana-sdk-example/pkg/generated/resource/foo"
	"github.com/radiohead/grafana-sdk-example/pkg/instrument"
)

// Validator validates incoming admission requests for Foo resource.
// It uses a cache for verifying name uniqueness of objects.
type Validator struct {
	cache *Cache
	syncm map[string]struct{}
	mmutx sync.Mutex
}

// NewValidator returns a new Validator which uses provided Cache.
func NewValidator(cache *Cache) *Validator {
	return &Validator{
		cache: cache,
		syncm: make(map[string]struct{}),
	}
}

// Validate handles incoming validating admission request.
func (v *Validator) Validate(ctx context.Context, req *resource.AdmissionRequest) error {
	ctx, logger, span := instrument.StartSpan(ctx, "foo.validator#Validate",
		"action", req.Action,
		"group", req.Group,
		"kind", req.Kind,
		"name", req.Object.StaticMetadata().Name,
	)
	defer span.End()

	logger.Debug("validating incoming request")

	foo, ok := req.Object.(*foo.Object)
	if !ok {
		err := errors.New("error parsing object")

		logger.Error(
			"error validating incoming request",
			"error", err,
		)

		return err
	}

	name := foo.Spec.Name

	logger.Debug(
		"checking that spec.Name is unique",
		"spec.name", name,
	)

	if err := v.tryLock(name); err != nil {
		logger.Error(
			"error validating incoming request",
			"error", err,
		)

		return err
	}
	defer v.unlock(name)

	if _, err := v.cache.Get(ctx, foo.StaticMeta.Namespace, name); err == nil {
		err := fmt.Errorf("an object with name '%s' already exists", name)

		logger.Error(
			"error validating incoming request",
			"error", err,
		)

		return err
	}

	return nil
}

func (v *Validator) tryLock(key string) error {
	v.mmutx.Lock()
	defer v.mmutx.Unlock()

	if _, ok := v.syncm[key]; ok {
		return fmt.Errorf("'%s' is already being validated", key)
	}

	v.syncm[key] = struct{}{}
	return nil
}

func (v *Validator) unlock(key string) {
	v.mmutx.Lock()
	delete(v.syncm, key)
	v.mmutx.Unlock()
}
