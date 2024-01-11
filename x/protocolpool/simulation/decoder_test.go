package simulation_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/protocolpool"
	"cosmossdk.io/x/protocolpool/simulation"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestDecodePoolStore(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(protocolpool.AppModuleBasic{})
	cdc := encCfg.Codec

	dec := simulation.NewDecodeStore(cdc)

	period := time.Duration(60) * time.Second
	fooCoin := sdk.NewInt64Coin("foo", 100)
	budget := types.Budget{
		RecipientAddress: sdk.AccAddress("cosmos1_______").String(),
		TotalBudget:      &fooCoin,
		StartTime:        &time.Time{},
		Tranches:         2,
		Period:           &period,
	}

	percentage, err := math.LegacyNewDecFromStr("0.2")
	require.NoError(t, err)
	cf := types.ContinuousFund{
		Recipient:  sdk.AccAddress("cosmos1a________").String(),
		Percentage: percentage,
		Expiry:     &time.Time{},
	}

	oneIntBz, err := math.OneInt().Marshal()
	require.NoError(t, err)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.BudgetKey, Value: cdc.MustMarshal(&budget)},
			{Key: types.ContinuousFundKey, Value: cdc.MustMarshal(&cf)},
			{Key: types.RecipientFundPercentageKey, Value: oneIntBz},
			{Key: types.RecipientFundDistributionKey, Value: oneIntBz},
			{Key: types.ToDistributeKey, Value: oneIntBz},
		},
	}

	tests := []struct {
		name   string
		expLog string
	}{
		{"Budget", fmt.Sprintf("%v\n%v", budget, budget)},
		{"ContinuousFund", fmt.Sprintf("%v\n%v", cf, cf)},
		{"RecipientFundPercentage", fmt.Sprintf("%v\n%v", math.OneInt(), math.OneInt())},
		{"RecipientFundDistribution", fmt.Sprintf("%v\n%v", math.OneInt(), math.OneInt())},
		{"ToDistribute", fmt.Sprintf("%v\n%v", math.OneInt(), math.OneInt())},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
