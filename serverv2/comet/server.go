package comet

import (
	"context"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

/*
items missing

- abciquery
	- queryApp
		- simulate
		- version
	- queryStore
		- key/value/proof (proof is optional)
	- queryP2P
		- filter peers
- store
- snapshots
*/

type Service interface {
	Start() error
	Stop() error
}

// TxCodec defines an interface that can be used to decode and encode transactions.
// Cosmos SDK is limited to protobuf encoded transactions today, the goal here to allow teams to define custom encoding,
// but the structure of the TX must stay the same
type TxCodec[T any] interface {
	Encode(tx T) ([]byte, error)
	Decode(txBytes []byte) (T, error)
}

// TxValidator defines an interface that can be used to validate transactions.
type TxValidator[T any] interface {
	// Validate validates a transaction
	Validate(tx T) error
	// AsyncValidate validates a transaction asynchronously
	AsyncValidate(txs []T) error
	// Cache caches a transaction if validation passed
	// this is used to avoid re-validating a transaction
	Cache(tx T) bool
}

type AppManager interface {
	Service
	Init() error
	DeliverTx(txs [][]byte) ([]byte, error)
}

type Params interface {
	GetParams() cmtproto.ConsensusParams
}

// VoteExtension is an interface that can be used to extend the voting process
// Note: needs access to branching of current states
type VoteExtension interface {
	ExtendVote(req any) (resp any, err error) // Extend Vote will be a handler set on the server
	VerifyVote(req any) (resp any, err error) // Verify Vote will be a handler set on the server
}

// Proposal is an interface that can be used to extend the proposal process and voting of the proposal
// Note: needs access to branching of current states
type Proposal interface {
	Prepare(req any) (resp any, err error) // Prepare will be a handler set on the server
	Process(req any) (resp any, err error) // Process will be a handler set on the server
}

type Mempool[T any] interface {
	Service
	Validate(context.Context, T) error
	ValidateAsync(context.Context, []T) error
	Select(context.Context, []T, uint32) ([]T, error)

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a set of transactiosn from the mempool in a single atomic operation.
	Remove([]byte) error
}

func newAppManager[T any]() (AppManager, TxCodec[T], TxValidator[T], error) {
	return nil, nil, nil, nil
}

func newMempool[T any](TxCodec[T], TxValidator[T]) (Mempool[T], error) {
	return nil, nil
}

type Options struct {
	ve VoteExtension
	p  Proposal
}

func newConsensus[T any](am AppManager, mempool Mempool[T], opts ...Options) (any, error) {
	return nil, nil
}

type Server[T any] struct {
	appManager  AppManager
	txCodec     TxCodec[T]
	txValidator TxValidator[T]
	mempool     Mempool[T]
	// consensus  consensus.Consensus
	// api        api.API
	// grpc 		 grpc.GRPC
}

func NewServer[T any]() (Server[T], error) {

	server := Server[T]{}
	// start app manager
	appmanager, txc, txv, err := newAppManager[T]()
	if err != nil {
		return Server[T]{}, err
	}
	server.appManager = appmanager
	server.txCodec = txc
	server.txValidator = txv

	// pass in the tx codec and txvalidator into the mempool ot handle valdiation and decoding
	mempool, err := newMempool(txc, txv)
	if err != nil {
		return Server[T]{}, err
	}
	server.mempool = mempool

	/*
		proposal := cometwrapper.NewProposal()
		voteextension := cometwrapper.NewVoteExtension()

	*/

	// consensus is a cometwrapper in this case that gets passed a mempool interface and app manager interface to handle txs
	// pass in votextension and proposal handlers into consensus
	_, err = newConsensus[T](appmanager, mempool)
	if err != nil {
		return Server[T]{}, err
	}
	// server.consensus = consensus

	// setup api & grpc servers

	return Server[T]{}, nil
}

func (s *Server[T]) Start() error {
	// appManager.Start()
	// consensus.Start()
	return nil
}

func (s *Server[T]) Stop() error {
	// appManager.Stop()
	// consensus.Stop()
	return nil
}
