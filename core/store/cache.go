package store

import "context"

// ObjectCacheService is a service that allows to store
// objects in a cache.
type ObjectCacheService interface {
	OpenObjectCache(ctx context.Context) ObjectCache
}

type ObjectCache interface {
	Cache(any, any)
	Get(any) (any, bool)
}
