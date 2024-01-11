package simulation

import (
	"bytes"
	"fmt"

	"cosmossdk.io/math"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding protocolpool type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.BudgetKey):
			var budgetA, budgetB types.Budget
			cdc.MustUnmarshal(kvA.Value, &budgetA)
			cdc.MustUnmarshal(kvB.Value, &budgetB)
			return fmt.Sprintf("%v\n%v", budgetA, budgetB)
		case bytes.Equal(kvA.Key[:1], types.ContinuousFundKey):
			var cfA, cfB types.ContinuousFund
			cdc.MustUnmarshal(kvA.Value, &cfA)
			cdc.MustUnmarshal(kvB.Value, &cfB)
			return fmt.Sprintf("%v\n%v", cfA, cfB)
		case bytes.Equal(kvA.Key[:1], types.RecipientFundPercentageKey):
			var rfpA, rfpB math.Int
			if err := rfpA.Unmarshal(kvA.Value); err != nil {
				panic(err)
			}
			if err := rfpB.Unmarshal(kvB.Value); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", rfpA, rfpB)
		case bytes.Equal(kvA.Key[:1], types.RecipientFundDistributionKey):
			var rfdA, rfdB math.Int
			if err := rfdA.Unmarshal(kvA.Value); err != nil {
				panic(err)
			}
			if err := rfdB.Unmarshal(kvB.Value); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", rfdA, rfdB)
		case bytes.Equal(kvA.Key[:1], types.ToDistributeKey):
			var toDistrA, toDistrB math.Int
			if err := toDistrA.Unmarshal(kvA.Value); err != nil {
				panic(err)
			}
			if err := toDistrB.Unmarshal(kvB.Value); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", toDistrA, toDistrB)
		default:
			panic(fmt.Sprintf("invalid protocolpool key %X", kvA.Key))
		}
	}
}
