package loader

import (
	"context"
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
	client   client.Client
	db       database.Database
	inputCh  chan *LoaderMsg[BlockRange]
	outputCh chan *LoaderStats

	ctx      context.Context
	stopOnce sync.Once
	g        *errgroup.Group
}

// LoaderOptions represents a map of arbitrary options when constructing a LoaderManager.
type LoaderOptions map[string]interface{}

type BlockRange struct {
	Start int64
	End   int64
}

// NewLoaderManager creates a new LoaderManager and initiates goroutines for the loaders in a pipeline.
func NewLoaderManager(ctx context.Context, client client.Client, db database.Database, opts LoaderOptions) (*LoaderManager, error) {
	// initialize struct
	g, ctx := errgroup.WithContext(ctx)
	loader := &LoaderManager{
		client:  client,
		db:      db,
		inputCh: make(chan *LoaderMsg[BlockRange]),
		ctx:     ctx,
		g:       g,
	}

	// loaders
	blockRangeLoader := NewLoader(client, loader.inputCh, blockRangeHandler)
	g.Go(blockRangeLoader.Run)
	blockHashLoader := NewLoader(client, blockRangeLoader.Dst(), blockHashHandler)
	g.Go(blockHashLoader.Run)
	blockLoader := NewLoaderSink(blockHashLoader.Dst(), blockHandler)
	g.Go(blockLoader.Run)

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
func (loader *LoaderManager) SendInput(blockRange BlockRange) (*LoaderStats, error) {
	dbTx, err := loader.db.NewDBTx()
	if err != nil {
		return nil, err
	}
	dbTxWithStats := newDBTxWithStats(dbTx)
	msg := &LoaderMsg[BlockRange]{
		dbTxWithStats,
		blockRange,
	}
	loader.inputCh <- msg
	return dbTxWithStats.LoaderStats, nil
}
