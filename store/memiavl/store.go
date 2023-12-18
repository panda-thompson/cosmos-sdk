package memiavl

import (
	"fmt"
	"io"
	"time"

	"github.com/crypto-org-chain/cronos/memiavl"
	abci "github.com/tendermint/tendermint/abci/types"

	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

var (
	_ types.KVStore                 = (*Store)(nil)
	_ types.CommitStore             = (*Store)(nil)
	_ types.CommitKVStore           = (*Store)(nil)
	_ types.Queryable               = (*Store)(nil)
	_ types.StoreWithInitialVersion = (*Store)(nil)
)

type Store struct {
	name string
	*memiavl.MultiTree
	*memiavl.Tree
}

func LoadStore(key types.StoreKey, multiTree *memiavl.MultiTree) (types.CommitKVStore, error) {
	tree := multiTree.TreeByName(key.Name())
	if tree == nil {
		return nil, fmt.Errorf("tree not found for key %s", key.Name())
	}
	return &Store{
		name:      key.Name(),
		MultiTree: multiTree,
		Tree:      tree,
	}, nil
}

func (s *Store) SetInitialVersion(version int64) {
	//TODO implement me
	panic("implement me")
}

func (s *Store) Query(query abci.RequestQuery) abci.ResponseQuery {
	//TODO implement me
	panic("implement me")
}

func (s *Store) Commit() types.CommitID {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "commit")
	hash, version, err := s.Tree.SaveVersion(true)
	if err != nil {
		panic(err)
	}
	return types.CommitID{
		Version: version,
		Hash:    hash,
	}
}

func (s *Store) LastCommitID() types.CommitID {
	return types.CommitID{
		Version: s.Tree.Version(),
		Hash:    s.Tree.RootHash(),
	}
}

func (s *Store) SetPruning(options pruningtypes.PruningOptions) {
	//TODO implement me
	panic("implement me")
}

func (s *Store) GetPruning() pruningtypes.PruningOptions {
	//TODO implement me
	panic("implement me")
}

func (s *Store) GetStoreType() types.StoreType {
	return types.StoreTypeIAVL
}

func (s *Store) CacheWrap() types.CacheWrap {
	return cachekv.NewStore(s)
}

func (s *Store) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(s, w, tc))
}

func (s *Store) CacheWrapWithListeners(storeKey types.StoreKey, listeners []types.WriteListener) types.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(s, storeKey, listeners))
}

func (s *Store) Get(key []byte) []byte {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "get")
	return s.Tree.Get(key)
}

func (s *Store) Has(key []byte) bool {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "has")
	return s.Tree.Has(key)
}

func (s *Store) Set(key, value []byte) {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "set")
	types.AssertValidKey(key)
	types.AssertValidValue(value)
	s.Tree.Set(key, value)
}

func (s *Store) Delete(key []byte) {
	defer telemetry.MeasureSince(time.Now(), "store", "iavl", "delete")
	s.Tree.Remove(key)
}

func (s *Store) Iterator(start, end []byte) types.Iterator {
	return s.Tree.Iterator(start, end, true)
}

func (s *Store) ReverseIterator(start, end []byte) types.Iterator {
	return s.Tree.Iterator(start, end, false)
}
