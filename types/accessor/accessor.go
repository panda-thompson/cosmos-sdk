package accessor

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var requestKey = &struct{}{}
var codecKey = &struct{}{}

// this errors will need to be update to use sdk/errors
var ErrNotFound = errors.New("context value not found")
var ErrWrongType = errors.New("context value has a wrong type")

/*
TOOD:
   + decide if we want to provide moduleKey through context, or module will manage it by himself
   + at
*/

func SetupContext(ctx context.Context, r *sdk.Request, cdc codec.Codec) context.Context {
	ctx = context.WithValue(ctx, requestKey, r)
	return context.WithValue(ctx, codecKey, cdc)
}

func Codec(ctx context.Context) (codec.Codec, error) {
	cdcI := ctx.Value(codecKey)
	if cdcI == nil {
		return nil, ErrNotFound
	}
	cdc, ok := cdcI.(codec.Codec)
	if !ok {
		return nil, ErrWrongType
	}
	return cdc, nil

}

func Store(ctx context.Context, moduleKey sdk.StoreKey) (ObjStore, error) {
	rI := ctx.Value(requestKey)
	if rI == nil {
		return nil, ErrNotFound
	}
	r, ok := rI.(*sdk.Request)
	if !ok {
		return nil, ErrWrongType
	}
	cdc, err := Codec(ctx)
	if err != nil {
		return nil, err
	}
	return objStore{r.KVStore(moduleKey), cdc}, nil
}

type ObjStore interface {
	sdk.KVStore

	Save(key []byte, v proto.Message) error
	MustSave(key []byte, v proto.Message)
}

type objStore struct {
	s   sdk.KVStore
	cdc codec.Codec
}

func (s objStore) Save(key []byte, obj proto.Message) error {
	bz, err := s.cdc.Marshal(obj)
	if err != nil {
		return err
	}
	return s.s.Set(key, bz)

}
