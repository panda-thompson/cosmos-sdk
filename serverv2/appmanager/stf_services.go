package appmanager

import (
	"context"

	store "cosmossdk.io/core/store"
)

type storeService struct{}

func (s storeService) OpenKVStore(ctx context.Context) store.KVStore {
	return ctx.(*executionContext).store
}

func NewStoreService() store.KVStoreService {
	return storeService{}
}
