package loader

import (
	"context"
	"fmt"
	"sync"

	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"golang.org/x/sync/errgroup"
)

// LoaderOptions represents a map of arbitrary options when constructing a LoaderManager.
type LoaderOptions map[string]interface{}

// ILoaderManager outlines an interface for a loader manager.
type ILoaderManager interface {
	SendInput(BlockRange, database.DBTx)
	Close()
}

// LoaderManager stores state for managing loaders for loading data from a `Client` to
// a `Database`.
type LoaderManager struct {
	client  client.Client
	db      database.Database
	inputCh chan *LoaderMsg[BlockRange]

	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once
	g        *errgroup.Group
}

// LoaderMsg stores state for data being passed through loaders.
type LoaderMsg[T any] struct {
	blockRange BlockRange
	dbTx       *database.DBTx
	data       T
}

type BlockRange struct {
	startBlockHeight int64
	endBlockHeight   int64
}

// NewLoaderManager creates a new LoaderManager and initiates goroutines for the loaders in a pipeline.
func NewLoaderManager(ctx context.Context, client client.Client, db database.Database, opts LoaderOptions) (*LoaderManager, error) {
	// initialize struct
	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)
	inputCh := make(chan *LoaderMsg[BlockRange])
	loader := &LoaderManager{
		client:  client,
		db:      db,
		inputCh: inputCh,
		ctx:     ctx,
		cancel:  cancel,
		g:       g,
	}

	// loaders
	blockRangeLoader := NewLoader[BlockRange, []*chainhash.Hash](client, inputCh, blockRangeHandler)
	g.Go(blockRangeLoader.Run)
	blockHashLoader := NewLoader[[]*chainhash.Hash, []*types.Block](client, blockRangeLoader.Dst(), blockHashHandler)
	g.Go(blockHashLoader.Run)
	blockLoader := NewLoaderSink[[]*types.Block](blockHashLoader.Dst(), blockHandler)
	g.Go(blockLoader.Run)

	return loader, nil
}

// Close gracefully shuts down all loader processes.
func (loader *LoaderManager) Close() error {
	done := false
	loader.stopOnce.Do(func() {
		close(loader.inputCh)
		loader.cancel()
		done = true
	})
	if !done {
		return fmt.Errorf("already closed")
	}
	return loader.g.Wait()
}

// SendInput starts the given parameters on the first stage of the loader pipeline.
func (loader *LoaderManager) SendInput(blockRange BlockRange, dbTx database.DBTx) {
	msg := &LoaderMsg[BlockRange]{
		blockRange,
		&dbTx,
		blockRange,
	}
	loader.inputCh <- msg
}

// ILoader is a simple interface for loaders.
type ILoader interface {
	Run() error
}

// Loader is the go to loader for an inidividual stage in the pipeline
// extracting data from an RPC client to a database. It stores state and
// uses the function f to transform data coming from a src channel to send
// to a dst channel.
type Loader[S, D any] struct {
	client client.Client
	src    <-chan *LoaderMsg[S]
	dst    chan *LoaderMsg[D]
	f      LoaderFunc[S, D]
}

type LoaderFunc[S, D any] func(client.Client, *LoaderMsg[S]) (*LoaderMsg[D], error)

func (loader *Loader[S, D]) Dst() <-chan *LoaderMsg[D] {
	return loader.dst
}

func NewLoader[S, D any](client client.Client, src <-chan *LoaderMsg[S], f LoaderFunc[S, D]) *Loader[S, D] {
	dst := make(chan *LoaderMsg[D])
	loader := &Loader[S, D]{
		client,
		src,
		dst,
		f,
	}
	return loader
}

// Run listens for messages sent to the loader's src channel,
// transforms the data and sends it to the next loader.
// Importantly, when it's src channel is closed, it closes its
// dst channel, which causes a domino effect of closing loader channels.
func (loader *Loader[S, D]) Run() error {
	for msg := range loader.src {
		output, err := loader.f(loader.client, msg)
		if err != nil {
			return err
		}
		loader.dst <- output
	}
	close(loader.dst)
	return nil
}

// LoaderSink represents the last stage in a loader pipeline.
type LoaderSink[S any] struct {
	src <-chan *LoaderMsg[S]
	f   LoaderSinkFunc[S]
}

type LoaderSinkFunc[S any] func(database.DBTx, *LoaderMsg[S]) error

func NewLoaderSink[S any](src <-chan *LoaderMsg[S], f LoaderSinkFunc[S]) *LoaderSink[S] {
	loader := &LoaderSink[S]{
		src,
		f,
	}
	return loader
}

func (loader *LoaderSink[S]) Run() error {
	for msg := range loader.src {
		err := loader.f(*msg.dbTx, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// blockRangeHandler is a LoaderFunc that uses a block range to
// retrieve a list of block hashes.
func blockRangeHandler(client client.Client, msg *LoaderMsg[BlockRange]) (*LoaderMsg[[]*chainhash.Hash], error) {
	blockRange := msg.data
	hashes, err := client.GetBlockHashesByRange(blockRange.startBlockHeight, blockRange.endBlockHeight)
	if err != nil {
		return nil, err
	}
	newMsg := &LoaderMsg[[]*chainhash.Hash]{msg.blockRange, msg.dbTx, hashes}
	return newMsg, nil
}

// blockHashHandler is a LoaderFunc that uses a list of block
// hashes to retrieve block header and transaction data.
func blockHashHandler(client client.Client, msg *LoaderMsg[[]*chainhash.Hash]) (*LoaderMsg[[]*types.Block], error) {
	hashes := msg.data
	blocks, err := client.GetBlocks(hashes)
	if err != nil {
		return nil, err
	}
	newMsg := &LoaderMsg[[]*types.Block]{msg.blockRange, msg.dbTx, blocks}
	return newMsg, nil
}

// blockHandler is a LoaderSinkFunc that fills a database transaction with
// block headers and transactions.
func blockHandler(dbTx database.DBTx, msg *LoaderMsg[[]*types.Block]) error {
	blocks := msg.data
	for _, block := range blocks {
		dbTx.AddBlockHeader(&block.BlockHeader)
		for _, tx := range block.Transactions() {
			dbTx.AddTransaction(tx)
		}
	}
	return nil
}
