package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Keeper struct {
	storeService storetypes.KVStoreService
	authKeeper   types.AccountKeeper
	bankKeeper   types.BankKeeper

	cdc codec.BinaryCodec

	authority string

	// State
	Schema         collections.Schema
	FundsDispensed collections.Map[sdk.AccAddress, sdk.Coins]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService,
	ak types.AccountKeeper, bk types.BankKeeper, authority string,
) Keeper {
	// ensure pool module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	sb := collections.NewSchemaBuilder(storeService)

	keeper := Keeper{
		storeService:   storeService,
		authKeeper:     ak,
		bankKeeper:     bk,
		cdc:            cdc,
		authority:      authority,
		FundsDispensed: collections.NewMap(sb, types.DispensableFundsKey, "dispensable_funds", sdk.AccAddressKey, codec.CollValue[sdk.Coins](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	keeper.Schema = schema

	return keeper
}

// GetAuthority returns the x/protocolpool module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With(log.ModuleKey, "x/"+types.ModuleName)
}

// FundCommunityPool allows an account to directly fund the community fund pool.
func (k Keeper) FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
}

// DistributeFromFeePool distributes funds from the protocolpool module account to
// a receiver address.
func (k Keeper) DistributeFromFeePool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
}

// GetCommunityPool get the community pool balance.
func (k Keeper) GetCommunityPool(ctx context.Context) (sdk.Coins, error) {
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAccount == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccount)
	}
	return k.bankKeeper.GetAllBalances(ctx, moduleAccount.GetAddress()), nil
}

func (k Keeper) GetTotalTaxFunds(ctx context.Context, recipient sdk.AccAddress) (sdk.Coins, error) {
	return k.FundsDispensed.Get(ctx, recipient)
	// amount, err := k.FundsDispensed.Get(ctx, recipient)
	// if err != nil {
	// 	return sdk.Coin{}, err
	// }
}

func (k Keeper) UpdateTotalTaxFunds(ctx context.Context, newFunds sdk.Coins, recipient sdk.AccAddress) error {
	// Get the current total tax or percentage-based funds
	currentFunds, err := k.GetTotalTaxFunds(ctx, recipient)
	if err != nil {
		return err
	}

	// Add the newly calculated funds to the current total
	totalFunds := currentFunds.Add(newFunds)

	// k.FundsDispensed.
}

func (k Keeper) isCapReached(ctx context.Context, cap sdk.Coins, recipient sdk.AccAddress) (bool, error) {
	// Get the total amount of funds moved so far (you may need to adjust this based on your implementation)
	totalMovedFunds, err := k.GetTotalFundsDispensed(ctx, recipient)
	if err != nil {
		return false, err
	}

	// Calculate the total amount that would be moved based on the provided cap
	proposedTotal := sdk.NewCoins(totalMovedFunds).Add(cap...)

	// Check if the proposed total exceeds the allowed cap
	cp, err := k.GetCommunityPool(ctx)
	if err != nil {
		return false, err
	}
	return proposedTotal.IsAnyGT(cp), nil

	// return false, nil
}
