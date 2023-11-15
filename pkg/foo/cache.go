package foo

import (
	"context"
	"errors"
	"sync"

	foo "github.com/radiohead/grafana-sdk-example/pkg/generated/resource/foo/v1alpha1"
	"github.com/radiohead/grafana-sdk-example/pkg/instrument"
)

// Cache is a simple in-memory cache of foo.Objects
// which supports fetching objects by spec.Name.
type Cache struct {
	idx map[fooKey]foo.Object
	mut sync.RWMutex
}

// NewCache returns a new cache of initial size initSize.
func NewCache(initSize int) *Cache {
	return &Cache{
		idx: make(map[fooKey]foo.Object, initSize),
	}
}

// Get returns an object from the cache using its namespace and spec.Name.
// If no object is present an error is returned instead.
func (c *Cache) Get(ctx context.Context, namespace, name string) (foo.Object, error) {
	_, logger, span := instrument.StartSpan(ctx, "foo.Cache#Get",
		"namespace", namespace,
		"name", name,
	)
	defer span.End()

	c.mut.RLock()
	defer c.mut.RUnlock()

	val, ok := c.idx[fooKey{
		namespace: namespace,
		name:      name,
	}]
	if !ok {
		logger.Debug("object not found in cache")
		return foo.Object{}, errors.New("object not found")
	}

	logger.Debug("found an object in cache")

	return val, nil
}

// Add adds a new object to the cache.
// If the object was already present in the cache it will be overwritten.
func (c *Cache) Add(ctx context.Context, obj foo.Object) {
	_, logger, span := instrument.StartSpan(ctx, "foo.Cache#Add")
	defer span.End()

	c.mut.Lock()
	defer c.mut.Unlock()

	c.idx[fooKey{
		namespace: obj.StaticMeta.Namespace,
		name:      obj.Spec.Name,
	}] = obj

	logger.Debug(
		"added an object to the cache",
		"namespace", obj.StaticMeta.Namespace,
		"name", obj.Spec.Name,
	)
}

// Delete deletes an object from the cache.
// It's a no-op if the object is not present in the cache.
func (c *Cache) Delete(ctx context.Context, obj foo.Object) {
	_, logger, span := instrument.StartSpan(ctx, "foo.Cache#Delete")
	defer span.End()

	c.mut.Lock()
	defer c.mut.Unlock()

	delete(c.idx, fooKey{
		namespace: obj.StaticMeta.Namespace,
		name:      obj.Spec.Name,
	})

	logger.Debug(
		"removed an object from the cache",
		"namespace", obj.StaticMeta.Namespace,
		"name", obj.Spec.Name,
	)
}

type fooKey struct {
	namespace string
	name      string
}
