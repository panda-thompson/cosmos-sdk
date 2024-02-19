package registry

import (
	"fmt"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/cosmos/cosmos-sdk/tests/integration/tx/registry/aaa"
	//_ "github.com/cosmos/cosmos-sdk/tests/integration/tx/registry/zzz"
)

func TestGogoFieldDescriptorWithDynamicPbMessage(t *testing.T) {
	ctx := aaa.SigningContextWithGogoRegisteredType
	err := ctx.Validate()
	require.NoError(t, err)

	fd, err := gogoproto.HybridResolver.FindFileByPath("testpb/test.proto")
	require.NoError(t, err)
	var md protoreflect.MessageDescriptor
	for i := 0; i < fd.Messages().Len(); i++ {
		if fd.Messages().Get(i).Name() == "SimpleSigner" {
			md = fd.Messages().Get(i)
		}
	}
	require.NotNil(t, md)

	dyanmicMsgType := dynamicpb.NewMessageType(md)
	fields := dyanmicMsgType.Descriptor().Fields()
	var field protoreflect.FieldDescriptor
	for i := 0; i < fields.Len(); i++ {
		field = fields.Get(i)
		fmt.Println(field.Name())
	}

	dyanmicMsg := dynamicpb.NewMessage(md)
	dyanmicMsg.Set(field, protoreflect.ValueOfString("foo"))
	signerBz, err := ctx.GetSigners(dyanmicMsg)
	require.NoError(t, err)
	fmt.Println(string(signerBz[0]))
}
