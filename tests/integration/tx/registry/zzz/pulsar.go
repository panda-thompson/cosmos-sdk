package zzz

import (
	"fmt"
	"testing"

	"cosmossdk.io/x/tx/signing"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	pulsartestpb "github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/pulsar/testpb"
)

func TestFailure(ctx *signing.Context, t *testing.T) {
	err := ctx.Validate()
	require.NoError(t, err)

	msg := &pulsartestpb.SimpleSigner{}
	md := msg.ProtoReflect().Descriptor()

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
