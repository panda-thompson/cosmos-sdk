package cometbft

import (
	"context"

	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/core/transaction"
	"cosmossdk.io/server/v2/stf/mock"
	abci "github.com/cometbft/cometbft/abci/types"
)

var _ abci.Application = (*Consensus[mock.Tx])(nil)

func NewConsensus[T transaction.Tx](app appmanager.AppManager[T]) *Consensus[T] {
	return &Consensus[T]{
		app: app,
	}
}

type Consensus[T transaction.Tx] struct {
	app appmanager.AppManager[T]
}

func (c Consensus[T]) Info(ctx context.Context, info *abci.RequestInfo) (*abci.ResponseInfo, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) Query(ctx context.Context, query *abci.RequestQuery) (*abci.ResponseQuery, error) {

}

func (c Consensus[T]) CheckTx(ctx context.Context, tx *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) InitChain(ctx context.Context, chain *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) PrepareProposal(ctx context.Context, proposal *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) ProcessProposal(ctx context.Context, proposal *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) FinalizeBlock(ctx context.Context, block *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) ExtendVote(ctx context.Context, vote *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) VerifyVoteExtension(ctx context.Context, extension *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) Commit(ctx context.Context, commit *abci.RequestCommit) (*abci.ResponseCommit, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) ListSnapshots(ctx context.Context, snapshots *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) OfferSnapshot(ctx context.Context, snapshot *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) LoadSnapshotChunk(ctx context.Context, chunk *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	// TODO implement me
	panic("implement me")
}

func (c Consensus[T]) ApplySnapshotChunk(ctx context.Context, chunk *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	// TODO implement me
	panic("implement me")
}
