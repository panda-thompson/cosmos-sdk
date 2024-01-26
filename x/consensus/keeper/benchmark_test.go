package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
	storetypes "cosmossdk.io/store/types"
	authtypes "cosmossdk.io/x/auth/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/require"
)

var sink any

var benchConsensusParams = cmtproto.ConsensusParams{
	Block: &cmtproto.BlockParams{
		MaxBytes: 500000,
		MaxGas:   5000,
	},
	Evidence: &cmtproto.EvidenceParams{
		MaxAgeNumBlocks: 54959854,
		MaxAgeDuration:  549548,
		MaxBytes:        549485,
	},
	Validator: &cmtproto.ValidatorParams{PubKeyTypes: []string{"cookies and cream"}},
	Version:   &cmtproto.VersionParams{App: 298459584},
	Abci: &cmtproto.ABCIParams{
		VoteExtensionsEnableHeight: 598549845985,
	},
}

func Benchmark_GettingParams(b *testing.B) {
	key := storetypes.NewKVStoreKey(StoreKey)
	testCtx := testutil.DefaultContextWithDB(b, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig()
	storeService := runtime.NewKVStoreService(key)

	keeper := NewKeeper(encCfg.Codec, storeService, authtypes.NewModuleAddress("gov").String(), runtime.EventService{})
	err := keeper.ParamsStore.Set(ctx, benchConsensusParams)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cp, _ := keeper.ParamsStore.Get(ctx)
		sink = cp
	}

	b.StopTimer()

	if sink == nil {
		b.Fatal("failed bench")
	}
}

func Benchmark_STFContext(b *testing.B) {
	db := branch.DefaultNewWriterMap(mock.DB())
	storeService := stf.NewStoreService([]byte("consensus"))

	s, err := stf.NewSTFBuilder[mock.Tx]().Build(&stf.STFBuilderOptions{})
	require.NoError(b, err)
	ctx := s.MakeContext(context.Background(), nil, db, 10000)

	keeper := NewKeeper(moduletestutil.MakeTestEncodingConfig().Codec, storeService, authtypes.NewModuleAddress("gov").String(), runtime.EventService{})
	err = keeper.ParamsStore.Set(ctx, benchConsensusParams)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cp, _ := keeper.ParamsStore.Get(ctx)
		sink = cp
	}

	b.StopTimer()

	if sink == nil {
		b.Fatal("failed bench")
	}
}
