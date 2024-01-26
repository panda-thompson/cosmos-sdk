package stf

import (
	"context"

	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/stf/objcache"
)

var _ store.KVStoreService = (*storeService)(nil)

func NewStoreService(address []byte) store.KVStoreService {
	return storeService{actor: address}
}

type storeService struct {
	actor []byte
}

func (s storeService) OpenKVStore(ctx context.Context) store.KVStore {
	state, err := ctx.(*executionContext).state.GetWriter(s.actor)
	if err != nil {
		panic(err)
	}
	return state
}

func (s storeService) OpenKeyLessContainer(ctx context.Context) objcache.KeylessContainer {
	return ctx.(*executionContext).objCache.GetKeylessContainer(s.actor)
}
func NewGasMeterService() gas.Service {
	return gasService{}
}

type gasService struct {
}

func (g gasService) GetGasMeter(ctx context.Context) gas.Meter {
	panic("impl")
}

func (g gasService) GetBlockGasMeter(ctx context.Context) gas.Meter {
	panic("stf has no block gas meter")
}

func (g gasService) WithGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	panic("impl")
}

func (g gasService) WithBlockGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	// TODO implement me
	panic("implement me")
}
