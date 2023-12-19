package serverv2_app

import "github.com/cosmos/cosmos-sdk/serverv2/core/transaction"

var _ transaction.Tx = Tx{}

// MOCK TX type.
type Tx struct {
}

func (t Tx) Hash() [32]byte { return [32]byte{} }

func (t Tx) GetMessages() []transaction.Type {
	// TODO implement me
	panic("implement me")
}

func (t Tx) GetSender() transaction.Identity {
	// TODO implement me
	panic("implement me")
}

func (t Tx) GetGasLimit() uint64 {
	// TODO implement me
	panic("implement me")
}
