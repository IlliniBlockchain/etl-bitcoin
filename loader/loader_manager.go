package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"golang.org/x/sync/errgroup"
)

// ILoaderManager outlines an interface for a loader manager.
type ILoaderManager interface {
	SendInput(BlockRange, database.DBTx)
	Close()
}

// LoaderManager stores state for managing loaders for loading data from a `Client` to
// a `Database`.
type LoaderManager struct {
	client  client.Client
	inputCh chan *LoaderMsg[BlockRange]
	// outputCh chan *LoaderStats

	ctx      context.Context
	stopOnce sync.Once
	g        *errgroup.Group
}

type BlockRange struct {
	Start int64
	End   int64
}

func (br BlockRange) String() string {
	return fmt.Sprintf("blocks %d - %d", br.Start, br.End)
}

// NewLoaderManager creates a new LoaderManager and initiates goroutines for the loaders in a pipeline.
func NewLoaderManager(ctx context.Context, client client.Client) (*LoaderManager, error) {
	// initialize struct
	g, ctx := errgroup.WithContext(ctx)
	loader := &LoaderManager{
		client:  client,
		inputCh: make(chan *LoaderMsg[BlockRange], 1),
		ctx:     ctx,
		g:       g,
	}

	// loaders
	blockRangeLoader := NewLoader(client, loader.inputCh, blockRangeHandler)
	blockHashLoader := NewLoader(client, blockRangeLoader.Dst(), blockHashHandler)
	blockLoader := NewLoaderSink(blockHashLoader.Dst(), blockHandler)
	loaders := []ILoader{blockRangeLoader, blockHashLoader, blockLoader}
	for _, loader := range loaders {
		l := loader
		g.Go(func() error {
			return l.Run(ctx)
		})
	}

	return loader, nil
}

// Close gracefully shuts down all loader processes.
func (loader *LoaderManager) Close() error {
	loader.stopOnce.Do(func() {
		close(loader.inputCh)
	})
	return loader.g.Wait()
}

// SendInput starts the given parameters on the first stage of the loader pipeline.
func (loader *LoaderManager) SendInput(blockRange BlockRange, dbTx database.DBTx) (*LoaderStats, error) {
	dbTxWithStats := newDBTxWithStats(dbTx, blockRange)
	msg := &LoaderMsg[BlockRange]{
		dbTxWithStats,
		blockRange,
	}
	select {
	case <-loader.ctx.Done():
		return nil, loader.ctx.Err()
	case loader.inputCh <- msg:
		return dbTxWithStats.LoaderStats, nil
	}
}
