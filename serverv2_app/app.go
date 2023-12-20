package serverv2_app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/serverv2/appmanager"
	appmanager_core "github.com/cosmos/cosmos-sdk/serverv2/core/appmanager"
	"github.com/cosmos/cosmos-sdk/serverv2_app/wirehack"
	"github.com/cosmos/cosmos-sdk/x/counter"
	counterkeeper "github.com/cosmos/cosmos-sdk/x/counter/keeper"
)

type SimApp struct {
	runtime   *appmanager.AppManager[Tx]
	consensus any //
}

func NewSTF() (*appmanager.STFAppManager[Tx], error) {
	builder := appmanager.NewSTFBuilder[Tx]()

	builder.AddModules(STFModules()...)

	return builder.Build(&appmanager.STFBuilderOptions{
		OrderEndBlockers:   nil,
		OrderBeginBlockers: nil,
		OrderTxValidators:  nil,
	})
}

func STFModules() []appmanager_core.STFModule[Tx] {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	// TODO: define stf store, define stf event manager
	counterKeeper := counterkeeper.NewKeeper(appmanager.NewStoreService(), nil)
	counterModule := counter.NewAppModule(counterKeeper)
	return wirehack.IntoSTFModules[Tx](cdc, counterModule)
}
