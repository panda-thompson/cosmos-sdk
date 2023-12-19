package wirehack

import (
	"context"

	cosmosmsg "cosmossdk.io/api/cosmos/msg/v1"
	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/serverv2/core/appmanager"
	"github.com/cosmos/cosmos-sdk/serverv2/core/transaction"
	"github.com/cosmos/cosmos-sdk/serverv2_app/wirehack/protocompat"
	"github.com/cosmos/cosmos-sdk/types/module"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoiface"
)

func IntoSTFModules[T transaction.Tx](cdc codec.Codec, appModules ...module.AppModule) []appmanager.STFModule[T] {
	var stfModules []appmanager.STFModule[T]
	for _, mod := range appModules {
		moduleWrapper := stfAppModuleWrapper[T]{
			appModule: mod,
			cdc:       cdc,
		}
		stfModules = append(stfModules, moduleWrapper)
	}
	return stfModules
}

var noOpBeginBlocker = func(_ context.Context) error { return nil }
var noOpEndBlocker = noOpBeginBlocker

type stfAppModuleWrapper[T transaction.Tx] struct {
	appModule module.AppModule
	cdc       codec.Codec
}

func (s stfAppModuleWrapper[T]) Name() string {
	return s.appModule.Name()
}

func (s stfAppModuleWrapper[T]) RegisterMsgHandlers(router appmanager.MsgRouterBuilder) {
	hasServices, ok := s.appModule.(appmodule.HasServices)
	if !ok {
		return
	}
	configurator := msgConfigurator{
		cdc:    s.cdc,
		router: router,
	}
	err := hasServices.RegisterServices(configurator)
	if err != nil {
		panic(err)
	}
}

func (s stfAppModuleWrapper[T]) RegisterQueryHandler(router appmanager.QueryRouterBuilder) {
	// TODO
	return
}

func (s stfAppModuleWrapper[T]) BeginBlocker() func(ctx context.Context) error {
	// check if has begin blockers
	beginBlocker, ok := s.appModule.(appmodule.HasBeginBlocker)
	if !ok {
		return noOpBeginBlocker
	}
	return beginBlocker.BeginBlock
}

func (s stfAppModuleWrapper[T]) EndBlocker() func(ctx context.Context) error {
	// check if has end blockers
	endBlocker, ok := s.appModule.(appmodule.HasEndBlocker)
	if !ok {
		return noOpEndBlocker
	}
	return endBlocker.EndBlock
}

func (s stfAppModuleWrapper[T]) UpdateValidators() func(ctx context.Context) ([]appmanager.ValidatorUpdate, error) {
	// look at this
	return func(ctx context.Context) ([]appmanager.ValidatorUpdate, error) {
		return nil, nil
	}
}

// here we should just be able to convert the T into an SDK.TX
func (s stfAppModuleWrapper[T]) TxValidator() func(ctx context.Context, tx T) error {
	return noOpTxValidator[T]()
}

func (s stfAppModuleWrapper[T]) RegisterPreMsgHandler(_ appmanager.PreMsgRouterBuilder) {}

func (s stfAppModuleWrapper[T]) RegisterPostMsgHandler(_ appmanager.PostMsgRouterBuilder) {}

func noOpTxValidator[T transaction.Tx]() func(ctx context.Context, tx T) error {
	return func(_ context.Context, _ T) error { return nil }
}

var _ module.Configurator = msgConfigurator{}

// this basically acts as module.Configurator, but converts the handlers into
// STF compatible ones
type msgConfigurator struct {
	cdc    codec.Codec
	err    error
	router appmanager.MsgRouterBuilder
}

func (m msgConfigurator) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	// register each method
	serviceDesc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(sd.ServiceName))
	if err != nil {
		panic(err)
	}
	if !protobuf.HasExtension(serviceDesc.(protoreflect.ServiceDescriptor).Options(), cosmosmsg.E_Service) {
		return
	}
	for _, md := range sd.Methods {
		err := m.registerMethod(sd, md, ss)
		if err != nil {
			panic(err)
		}
	}
}

func (m msgConfigurator) Error() error {
	return nil
}

func (m msgConfigurator) MsgServer() grpc1.Server {
	return m
}

func (m msgConfigurator) QueryServer() grpc1.Server {
	return noOpGRPC{}
}

func (m msgConfigurator) RegisterMigration(moduleName string, fromVersion uint64, handler module.MigrationHandler) error {
	// TODO: this is todo
	return nil
}

func (m msgConfigurator) registerMethod(sd *grpc.ServiceDesc, md grpc.MethodDesc, ss interface{}) error {
	requestName, err := protocompat.RequestFullNameFromMethodDesc(sd, md)
	if err != nil {
		return err
	}

	responseName, err := protocompat.ResponseFullNameFromMethodDesc(sd, md)
	if err != nil {
		return err
	}

	// now we create the hybrid handler
	handler, err := protocompat.MakeHybridHandler(m.cdc, sd, md, ss)
	if err != nil {
		return err
	}

	// STF likes to have only protov2
	requestV2Type, err := protoregistry.GlobalTypes.FindMessageByName(requestName)
	if err != nil {
		return err
	}

	responseV2Type, err := protoregistry.GlobalTypes.FindMessageByName(responseName)
	if err != nil {
		return err
	}

	m.router.RegisterHandler(requestV2Type.New().Interface(), func(ctx context.Context, msg appmanager.Type) (resp appmanager.Type, err error) {
		resp = responseV2Type.New().Interface()
		return resp, handler(ctx, msg.(protoiface.MessageV1), resp.(protoiface.MessageV1))
	})
	return nil
}

type noOpGRPC struct{}

func (n noOpGRPC) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {}
