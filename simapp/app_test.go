package simapp

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"gotest.tools/assert"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/auth"
	authsigning "cosmossdk.io/x/auth/signing"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/auth/vesting"
	"cosmossdk.io/x/authz"
	authzmodule "cosmossdk.io/x/authz/module"
	"cosmossdk.io/x/bank"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/distribution"
	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/feegrant"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/gov"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"
	group "cosmossdk.io/x/group/module"
	"cosmossdk.io/x/mint"
	"cosmossdk.io/x/protocolpool"
	"cosmossdk.io/x/slashing"
	"cosmossdk.io/x/staking"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	govv1beta1 "cosmossdk.io/api/cosmos/gov/v1beta1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func TestSimAppExportAndBlockedAddrs(t *testing.T) {
	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	app := NewSimappWithCustomOptions(t, false, SetupOptions{
		Logger:  logger.With("instance", "first"),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	})

	// BlockedAddresses returns a map of addresses in app v1 and a map of modules name in app v2.
	for acc := range BlockedAddresses() {
		var addr sdk.AccAddress
		if modAddr, err := sdk.AccAddressFromBech32(acc); err == nil {
			addr = modAddr
		} else {
			addr = app.AuthKeeper.GetModuleAddress(acc)
		}

		require.True(
			t,
			app.BankKeeper.BlockedAddr(addr),
			fmt.Sprintf("ensure that blocked addresses are properly set in bank keeper: %s should be blocked", acc),
		)
	}

	// finalize block so we have CheckTx state set
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err)

	_, err = app.Commit()
	require.NoError(t, err)

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewSimApp(logger.With("instance", "second"), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func TestRunMigrations(t *testing.T) {
	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	app := NewSimApp(logger.With("instance", "simapp"), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))

	// Create a new baseapp and configurator for the purpose of this test.
	bApp := baseapp.NewBaseApp(app.Name(), logger.With("instance", "baseapp"), db, app.TxConfig().TxDecoder())
	bApp.SetCommitMultiStoreTracer(nil)
	bApp.SetInterfaceRegistry(app.InterfaceRegistry())
	app.BaseApp = bApp
	configurator := module.NewConfigurator(app.appCodec, bApp.MsgServiceRouter(), app.GRPCQueryRouter())

	// We register all modules on the Configurator, except x/bank. x/bank will
	// serve as the test subject on which we run the migration tests.
	//
	// The loop below is the same as calling `RegisterServices` on
	// ModuleManager, except that we skip x/bank.
	for name, mod := range app.ModuleManager.Modules {
		if name == banktypes.ModuleName {
			continue
		}

		if mod, ok := mod.(module.HasServices); ok {
			mod.RegisterServices(configurator)
		}

		if mod, ok := mod.(appmodule.HasServices); ok {
			err := mod.RegisterServices(configurator)
			require.NoError(t, err)
		}

		require.NoError(t, configurator.Error())
	}

	// Initialize the chain
	_, err := app.InitChain(&abci.RequestInitChain{})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	testCases := []struct {
		name         string
		moduleName   string
		fromVersion  uint64
		toVersion    uint64
		expRegErr    bool // errors while registering migration
		expRegErrMsg string
		expRunErr    bool // errors while running migration
		expRunErrMsg string
		expCalled    int
	}{
		{
			"cannot register migration for version 0",
			"bank", 0, 1,
			true, "module migration versions should start at 1: invalid version", false, "", 0,
		},
		{
			"throws error on RunMigrations if no migration registered for bank",
			"", 1, 2,
			false, "", true, "no migrations found for module bank: not found", 0,
		},
		{
			"can register 1->2 migration handler for x/bank, cannot run migration",
			"bank", 1, 2,
			false, "", true, "no migration found for module bank from version 2 to version 3: not found", 0,
		},
		{
			"can register 2->3 migration handler for x/bank, can run migration",
			"bank", 2, bank.AppModule{}.ConsensusVersion(),
			false, "", false, "", int(bank.AppModule{}.ConsensusVersion() - 2), // minus 2 because 1-2 is run in the previous test case.
		},
		{
			"cannot register migration handler for same module & fromVersion",
			"bank", 1, 2,
			true, "another migration for module bank and version 1 already exists: internal logic error", false, "", 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			var err error

			// Since it's very hard to test actual in-place store migrations in
			// tests (due to the difficulty of maintaining multiple versions of a
			// module), we're just testing here that the migration logic is
			// called.
			called := 0

			if tc.moduleName != "" {
				for i := tc.fromVersion; i < tc.toVersion; i++ {
					// Register migration for module from version `fromVersion` to `fromVersion+1`.
					tt.Logf("Registering migration for %q v%d", tc.moduleName, i)
					err = configurator.RegisterMigration(tc.moduleName, i, func(sdk.Context) error {
						called++

						return nil
					})

					if tc.expRegErr {
						require.EqualError(tt, err, tc.expRegErrMsg)

						return
					}
					require.NoError(tt, err, "registering migration")
				}
			}

			// Run migrations only for bank. That's why we put the initial
			// version for bank as 1, and for all other modules, we put as
			// their latest ConsensusVersion.
			_, err = app.ModuleManager.RunMigrations(
				app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()}), configurator,
				module.VersionMap{
					"accounts":     accounts.AppModule{}.ConsensusVersion(),
					"bank":         1,
					"auth":         auth.AppModule{}.ConsensusVersion(),
					"authz":        authzmodule.AppModule{}.ConsensusVersion(),
					"staking":      staking.AppModule{}.ConsensusVersion(),
					"mint":         mint.AppModule{}.ConsensusVersion(),
					"distribution": distribution.AppModule{}.ConsensusVersion(),
					"slashing":     slashing.AppModule{}.ConsensusVersion(),
					"gov":          gov.AppModule{}.ConsensusVersion(),
					"group":        group.AppModule{}.ConsensusVersion(),
					"upgrade":      upgrade.AppModule{}.ConsensusVersion(),
					"vesting":      vesting.AppModule{}.ConsensusVersion(),
					"feegrant":     feegrantmodule.AppModule{}.ConsensusVersion(),
					"evidence":     evidence.AppModule{}.ConsensusVersion(),
					"genutil":      genutil.AppModule{}.ConsensusVersion(),
					"protocolpool": protocolpool.AppModule{}.ConsensusVersion(),
				},
			)
			if tc.expRunErr {
				require.EqualError(tt, err, tc.expRunErrMsg, "running migration")
			} else {
				require.NoError(tt, err, "running migration")
				// Make sure bank's migration is called.
				require.Equal(tt, tc.expCalled, called)
			}
		})
	}
}

func TestInitGenesisOnMigration(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewSimApp(log.NewTestLogger(t), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})

	// Create a mock module. This module will serve as the new module we're
	// adding during a migration.
	mockCtrl := gomock.NewController(t)
	t.Cleanup(mockCtrl.Finish)
	mockModule := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockDefaultGenesis := json.RawMessage(`{"key": "value"}`)
	mockModule.EXPECT().DefaultGenesis(gomock.Eq(app.appCodec)).Times(1).Return(mockDefaultGenesis)
	mockModule.EXPECT().InitGenesis(gomock.Eq(ctx), gomock.Eq(app.appCodec), gomock.Eq(mockDefaultGenesis)).Times(1)
	mockModule.EXPECT().ConsensusVersion().Times(1).Return(uint64(0))

	app.ModuleManager.Modules["mock"] = mockModule

	// Run migrations only for "mock" module. We exclude it from
	// the VersionMap to simulate upgrading with a new module.
	_, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(),
		module.VersionMap{
			"bank":         bank.AppModule{}.ConsensusVersion(),
			"auth":         auth.AppModule{}.ConsensusVersion(),
			"authz":        authzmodule.AppModule{}.ConsensusVersion(),
			"staking":      staking.AppModule{}.ConsensusVersion(),
			"mint":         mint.AppModule{}.ConsensusVersion(),
			"distribution": distribution.AppModule{}.ConsensusVersion(),
			"slashing":     slashing.AppModule{}.ConsensusVersion(),
			"gov":          gov.AppModule{}.ConsensusVersion(),
			"upgrade":      upgrade.AppModule{}.ConsensusVersion(),
			"vesting":      vesting.AppModule{}.ConsensusVersion(),
			"feegrant":     feegrantmodule.AppModule{}.ConsensusVersion(),
			"evidence":     evidence.AppModule{}.ConsensusVersion(),
			"genutil":      genutil.AppModule{}.ConsensusVersion(),
		},
	)
	require.NoError(t, err)
}

func TestUpgradeStateOnGenesis(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewSimappWithCustomOptions(t, false, SetupOptions{
		Logger:  log.NewTestLogger(t),
		DB:      db,
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	})

	// make sure the upgrade keeper has version map in state
	ctx := app.NewContext(false)
	vm, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	for v, i := range app.ModuleManager.Modules {
		if i, ok := i.(module.HasConsensusVersion); ok {
			require.Equal(t, vm[v], i.ConsensusVersion())
		}
	}
}

// TestMergedRegistry tests that fetching the gogo/protov2 merged registry
// doesn't fail after loading all file descriptors.
func TestMergedRegistry(t *testing.T) {
	r, err := proto.MergedRegistry()
	require.NoError(t, err)
	require.Greater(t, r.NumFiles(), 0)
}

func TestProtoAnnotations(t *testing.T) {
	r, err := proto.MergedRegistry()
	require.NoError(t, err)
	err = msgservice.ValidateProtoAnnotations(r)
	require.NoError(t, err)
}

var _ address.Codec = (*customAddressCodec)(nil)

type customAddressCodec struct{}

func (c customAddressCodec) StringToBytes(text string) ([]byte, error) {
	return []byte(text), nil
}

func (c customAddressCodec) BytesToString(bz []byte) (string, error) {
	return string(bz), nil
}

func TestAddressCodecFactory(t *testing.T) {
	var addrCodec address.Codec
	var valAddressCodec runtime.ValidatorAddressCodec
	var consAddressCodec runtime.ConsensusAddressCodec

	err := depinject.Inject(
		depinject.Configs(
			network.MinimumAppConfig(),
			depinject.Supply(log.NewNopLogger()),
		),
		&addrCodec, &valAddressCodec, &consAddressCodec)
	require.NoError(t, err)
	require.NotNil(t, addrCodec)
	_, ok := addrCodec.(customAddressCodec)
	require.False(t, ok)
	require.NotNil(t, valAddressCodec)
	_, ok = valAddressCodec.(customAddressCodec)
	require.False(t, ok)
	require.NotNil(t, consAddressCodec)
	_, ok = consAddressCodec.(customAddressCodec)
	require.False(t, ok)

	// Set the address codec to the custom one
	err = depinject.Inject(
		depinject.Configs(
			network.MinimumAppConfig(),
			depinject.Supply(
				log.NewNopLogger(),
				func() address.Codec { return customAddressCodec{} },
				func() runtime.ValidatorAddressCodec { return customAddressCodec{} },
				func() runtime.ConsensusAddressCodec { return customAddressCodec{} },
			),
		),
		&addrCodec, &valAddressCodec, &consAddressCodec)
	require.NoError(t, err)
	require.NotNil(t, addrCodec)
	_, ok = addrCodec.(customAddressCodec)
	require.True(t, ok)
	require.NotNil(t, valAddressCodec)
	_, ok = valAddressCodec.(customAddressCodec)
	require.True(t, ok)
	require.NotNil(t, consAddressCodec)
	_, ok = consAddressCodec.(customAddressCodec)
	require.True(t, ok)
}

func TestCustom(t *testing.T) {
	db := dbm.NewMemDB()
	app := NewSimApp(log.NewTestLogger(t), db, nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight()})

	app.BankKeeper.SetDenomMetaData(
		ctx,
		banktypes.Metadata{
			Description: "The native staking token of the Cosmos Hub.",
			Base:        "uatom",
			Display:     "ATOM",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "ATOM",
					Exponent: 6,
				},
			},
		},
	)

	// the MsgGrantAllowance message
	allowanceAny, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{SpendLimit: sdk.NewCoins(sdk.NewCoin("foo", math.NewInt(100)))})
	require.NoError(t, err)
	feegrantmsg := &feegrant.MsgGrantAllowance{Granter: "cosmos1qperwt9wrnkg5k9e5gzfgjppzpq0ah2h2l6mc5", Grantee: "cosmos1yeckxz7tapz34kjwnjxvmxzurerquhtrmx5j6g", Allowance: allowanceAny}

	// the MsgSubmitProposal message with a TextProposal
	newProposalMsg, _ := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000)},
		"cosmos1qperwt9wrnkg5k9e5gzfgjppzpq0ah2h2l6mc5",
		"",
		"Proposal",
		"description of proposal",
		v1.ProposalType_PROPOSAL_TYPE_STANDARD,
	)

	//authz grant
	m := &authz.MsgGrant{
		Granter: "cosmos1qperwt9wrnkg5k9e5gzfgjppzpq0ah2h2l6mc5",
		Grantee: "cosmos1yeckxz7tapz34kjwnjxvmxzurerquhtrmx5j6g",
	}
	g := authz.GenericAuthorization{Msg: "some_type_of_msg/MsgSomeTypeOfMsg"}
	m.Grant.Authorization, _ = codectypes.NewAnyWithValue(&g)

	msgs := []proto.Message{
		&bankv1beta1.MsgSend{
			FromAddress: "cosmos1qperwt9wrnkg5k9e5gzfgjppzpq0ah2h2l6mc5",
			ToAddress:   "cosmos1yeckxz7tapz34kjwnjxvmxzurerquhtrmx5j6g",
			Amount: []*basev1beta1.Coin{{
				Denom:  "uatom",
				Amount: "1000000",
			}},
		},
		&govv1beta1.MsgVote{
			ProposalId: 1,
			Voter:      "cosmos1qperwt9wrnkg5k9e5gzfgjppzpq0ah2h2l6mc5",
			Option:     govv1beta1.VoteOption_VOTE_OPTION_YES,
		},
		newProposalMsg,
		feegrantmsg,
		m,
	}

	txBuilder := app.TxConfig().NewTxBuilder()
	txBuilder.SetMsgs(msgs...)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("uatom", math.NewInt(1000))))
	txBuilder.SetGasLimit(200000)
	txBuilder.SetMemo("test")
	tx := txBuilder.GetTx()

	_, pubkey, _ := testdata.KeyTestPubAddr()
	anyPk, _ := codectypes.NewAnyWithValue(pubkey)

	signerData := txsigning.SignerData{
		Address:       "cosmos1qperwt9wrnkg5k9e5gzfgjppzpq0ah2h2l6mc5",
		ChainID:       "test-chain",
		AccountNumber: 3,
		Sequence:      123,
		PubKey: &anypb.Any{
			TypeUrl: anyPk.TypeUrl,
			Value:   anyPk.Value,
		},
	}
	adaptableTx, ok := tx.(authsigning.V2AdaptableTx)
	if !ok {
		panic("tx does not implement V2AdaptableTx")
	}
	txData := adaptableTx.GetSigningTxData()

	signbytes, err := app.TxConfig().SignModeHandler().GetSignBytes(
		ctx,
		signingv1beta1.SignMode_SIGN_MODE_TEXTUAL,
		signerData,
		txData,
	)

	require.NoError(t, err)

	panic(hex.EncodeToString(signbytes))

	/*
		a101982ca20168436861696e206964026a746573742d636861696ea2016e4163636f756e74206e756d626572026133a2016853657175656e63650263313233a301674164647265737302782d636f736d6f73317170657277743977726e6b67356b396535677a66676a70707a70713061683268326c366d633504f5a3016a5075626c6963206b657902781f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657904f5a401634b657902785230323443204538324620314538462039343245204541413920363543332043363741204144413920423031412035383841203139333120424430442030324235203636324520353543452035313941204134030104f5a102781f54686973207472616e73616374696f6e206861732035204d65737361676573a3016d4d6573736167652028312f352902781c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e640301a3016c46726f6d206164647265737302782d636f736d6f73317170657277743977726e6b67356b396535677a66676a70707a70713061683268326c366d63350302a3016a546f206164647265737302782d636f736d6f73317965636b787a377461707a33346b6a776e6a78766d787a7572657271756874726d78356a36670302a30166416d6f756e74026f312730303027303030207561746f6d0302a3016d4d6573736167652028322f352902781b2f636f736d6f732e676f762e763162657461312e4d7367566f74650301a3016b50726f706f73616c2069640261310302a30165566f74657202782d636f736d6f73317170657277743977726e6b67356b396535677a66676a70707a70713061683268326c366d63350302a301664f7074696f6e026f564f54455f4f5054494f4e5f5945530302a3016d4d6573736167652028332f35290278202f636f736d6f732e676f762e76312e4d73675375626d697450726f706f73616c0301a301684d6573736167657302653120416e790302a3016e4d657373616765732028312f31290278232f636f736d6f732e676f762e76312e4d7367457865634c6567616379436f6e74656e740303a30167436f6e74656e740278202f636f736d6f732e676f762e763162657461312e5465787450726f706f73616c0304a301655469746c650264546573740305a3016b4465736372697074696f6e026b6465736372697074696f6e0305a30169417574686f7269747902782d636f736d6f73313064303779323635676d6d757674347a30773961773838306a6e73723730306a367a6e396b6e0304a2026f456e64206f66204d657373616765730302a3016f496e697469616c206465706f736974026d31303027303030207374616b650302a3016850726f706f73657202782d636f736d6f73317170657277743977726e6b67356b396535677a66676a70707a70713061683268326c366d63350302a301655469746c65026850726f706f73616c0302a3016753756d6d61727902776465736372697074696f6e206f662070726f706f73616c0302a3016d50726f706f73616c2074797065027650524f504f53414c5f545950455f5354414e444152440302a3016d4d6573736167652028342f352902782a2f636f736d6f732e6665656772616e742e763162657461312e4d73674772616e74416c6c6f77616e63650301a301674772616e74657202782d636f736d6f73317170657277743977726e6b67356b396535677a66676a70707a70713061683268326c366d63350302a301674772616e74656502782d636f736d6f73317965636b787a377461707a33346b6a776e6a78766d787a7572657271756874726d78356a36670302a30169416c6c6f77616e63650278272f636f736d6f732e6665656772616e742e763162657461312e4261736963416c6c6f77616e63650302a3016b5370656e64206c696d6974026731303020666f6f0303a3016d4d6573736167652028352f352902781e2f636f736d6f732e617574687a2e763162657461312e4d73674772616e740301a301674772616e74657202782d636f736d6f73317170657277743977726e6b67356b396535677a66676a70707a70713061683268326c366d63350302a301674772616e74656502782d636f736d6f73317965636b787a377461707a33346b6a776e6a78766d787a7572657271756874726d78356a36670302a301654772616e74026c4772616e74206f626a6563740302a3016d417574686f72697a6174696f6e02782a2f636f736d6f732e617574687a2e763162657461312e47656e65726963417574686f72697a6174696f6e0303a301634d7367027821736f6d655f747970655f6f665f6d73672f4d7367536f6d65547970654f664d73670304a1026e456e64206f66204d657373616765a201644d656d6f026474657374a2016446656573026b3127303030207561746f6da30169476173206c696d697402673230302730303004f5a3017148617368206f66207261772062797465730278403763363034666532616434393563316164656566356565376331386265646138396162643134343062323661303137656465646535616431366237363462623004f5
	*/

}

// mkTestLegacyContent creates a MsgExecLegacyContent for testing purposes.
func mkTestLegacyContent(t *testing.T) *v1.MsgExecLegacyContent {
	t.Helper()
	TestProposal := v1beta1.NewTextProposal("Test", "description")
	msgContent, err := v1.NewLegacyContent(TestProposal, authtypes.NewModuleAddress(types.ModuleName).String())
	assert.NilError(t, err)

	return msgContent
}
