package types

import abci "github.com/tendermint/tendermint/abci/types"

type InitChainer func(ctx Context, req abci.RequestInitChain) abci.ResponseInitChain

type BeginBlocker func(ctx Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock

type EndBlocker func(ctx Context, req abci.RequestEndBlock) abci.ResponseEndBlock

type PeerFilter func(info string) abci.ResponseQuery
