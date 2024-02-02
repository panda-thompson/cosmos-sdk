package pruning

import (
	"context"
	"errors"
	"sync"

	"cosmossdk.io/log"
	"cosmossdk.io/store/v2"
)

func NewManagerV2(sc store.Committer, scOpts Options, ss store.VersionedDatabase, ssOpts Options, logger log.Logger) *ManagerV2 {
	return &ManagerV2{
		scOpts:   scOpts,
		scPruner: newPruner(sc, logger.With("state-commitment")),
		ssOpts:   ssOpts,
		ssPruner: newPruner(ss, logger.With("state-storage")),
	}
}

type ManagerV2 struct {
	startOnce sync.Once
	stopOnce  sync.Once

	scOpts   Options
	scPruner pruner

	ssOpts   Options
	ssPruner pruner
}

func (m *ManagerV2) Prune(h uint64) {
	if m.ssOpts.ShouldPrune(h) {
		m.ssPruner.prune(h)
	}
	if m.scOpts.ShouldPrune(h) {
		m.scPruner.prune(h)
	}
}

func (m *ManagerV2) Start() {
	m.startOnce.Do(func() {
		go m.ssPruner.loop()
		go m.scPruner.loop()
	})
}

func (m *ManagerV2) Stop(ctx context.Context) (err error) {
	// all these APIs should take a context
	m.stopOnce.Do(func() {
		err = errors.Join(m.ssPruner.close(ctx))
		err = errors.Join(m.scPruner.close(ctx))
	})
	return
}

// prunable generically defines something that can be pruned.
type prunable interface {
	Prune(height uint64) error
}

func newPruner(p prunable, logger log.Logger) pruner {
	return pruner{
		p:        p,
		pruneReq: make(chan uint64),
		closed:   make(chan struct{}),
		done:     make(chan struct{}),
		log:      logger,
	}
}

type pruner struct {
	p prunable

	pruneReq chan uint64   // receives new prune requests
	closed   chan struct{} // used by the parent to signal closure
	done     chan struct{} // closed when loop exits

	log log.Logger
}

func (p pruner) prune(height uint64) {
	select {
	case p.pruneReq <- height:
	default:
		p.log.Info("skipping pruning, busy", "height", height)
	}
}

func (p pruner) close(ctx context.Context) error {
	close(p.closed)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return nil
	}
}

func (p pruner) loop() {
	defer close(p.done)
	for {
		select {
		case <-p.done:
			return
		case h := <-p.pruneReq:
			p.internalPrune(h)
		}
	}
}

func (p pruner) internalPrune(height uint64) {
	err := p.p.Prune(height)
	if err != nil {
		p.log.Error("unable to prune", "height", height, "err", err)
		return
	}

	p.log.Debug("pruned height", "height", height)
}
