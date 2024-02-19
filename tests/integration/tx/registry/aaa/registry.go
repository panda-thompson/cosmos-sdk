package aaa

import (
	"encoding/hex"
	"strings"

	txsigning "cosmossdk.io/x/tx/signing"
	gogoproto "github.com/cosmos/gogoproto/proto"
	// critical to import the gogo/testpb package to register the test.proto file in the gogoproto registry
	_ "github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/gogo/testpb"
)

type dummyAddressCodec struct{}

func (d dummyAddressCodec) StringToBytes(text string) ([]byte, error) {
	return []byte(text), nil
}

func (d dummyAddressCodec) BytesToString(bz []byte) (string, error) {
	return string(bz), nil
}

type dummyValidatorAddressCodec struct{}

func (d dummyValidatorAddressCodec) StringToBytes(text string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(text, "val"))
}

func (d dummyValidatorAddressCodec) BytesToString(bz []byte) (string, error) {
	return "val" + hex.EncodeToString(bz), nil
}

func makeSigningContext() (*txsigning.Context, error) {
	ctx, err := txsigning.NewContext(txsigning.Options{
		AddressCodec:          dummyAddressCodec{},
		ValidatorAddressCodec: dummyValidatorAddressCodec{},
		FileResolver:          gogoproto.HybridResolver,
	})
	if err != nil {
		return nil, err
	}
	err = ctx.Validate()
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

var SigningContextWithGogoRegisteredType, ContextErr = makeSigningContext()
