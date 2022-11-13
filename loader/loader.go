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

// General idea for loader:
// Represents a single entire process of getting data from a source to a destination.
// It takes in clients (RPC or DB) and implements the loading process for transporting
// data between them.

// LoaderOptions represents a map of arbitrary options when constructing a loader.
type LoaderOptions map[string]interface{}

// ClientDatabaseLoader represents any process of loading data from a
// Client to a Database.
type Loader interface {
	SendInput(BlockRange, database.DBTx)
	Close()
}

type ClientDBLoaderConstructor func(client client.Client, db database.Database, opts LoaderOptions) (Loader, error)

type ClientDBLoader struct {
	client   client.Client
	db       database.Database
	pipeline ClientDBPipeline

	dbTxMu   sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once
	g        *errgroup.Group
}

type ClientDBPipeline struct {
	blockRangeCh      chan LoaderMsg
	blockHashesCh     chan LoaderMsg
	blocksToHeadersCh chan LoaderMsg
	blocksToTxsCh     chan LoaderMsg
	blockHeadersCh    chan LoaderMsg
	txsCh             chan LoaderMsg
}

func (pipeline *ClientDBPipeline) Close() {
	close(pipeline.blockRangeCh)
	close(pipeline.blockHashesCh)
	close(pipeline.blocksToHeadersCh)
	close(pipeline.blocksToTxsCh)
	close(pipeline.blockHeadersCh)
	close(pipeline.txsCh)
}

type LoaderMsg struct {
	blockRange BlockRange
	dbTx       *database.DBTx
	data       interface{}
}

// Message type enum
type LoaderMsgType int

const (
	Range LoaderMsgType = iota
	Hashes
	Blocks
	BlocksForHeaders
	BlocksForTxs
	Headers
	Txs
)

type BlockRange struct {
	startBlockHeight int64
	endBlockHeight   int64
}

func NewClientDBLoader(ctx context.Context, client client.Client, db database.Database, opts LoaderOptions) (*ClientDBLoader, error) {
	// initialize struct
	pipeline := ClientDBPipeline{}
	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)
	loader := &ClientDBLoader{
		client:   client,
		db:       db,
		pipeline: pipeline,
		ctx:      ctx,
		cancel:   cancel,
		g:        g,
	}

	// start a worker for each stage of pipeline
	g.Go(func() error {
		return loader.blockRangeHandler(loader.pipeline.blockRangeCh, loader.pipeline.blockHashesCh)
	})
	g.Go(func() error {
		return loader.blockHashesHandler(loader.pipeline.blockHashesCh, loader.pipeline.blocksToHeadersCh, loader.pipeline.blocksToTxsCh)
	})
	g.Go(func() error {
		return loader.blocksToHeadersHandler(loader.pipeline.blocksToHeadersCh, loader.pipeline.blockHeadersCh)
	})
	g.Go(func() error { return loader.blocksToTxsHandler(loader.pipeline.blocksToTxsCh, loader.pipeline.txsCh) })
	g.Go(func() error { return loader.headersHandler(loader.pipeline.blockHeadersCh) })
	g.Go(func() error { return loader.txsHandler(loader.pipeline.txsCh) })

	return loader, nil
}

func (loader *ClientDBLoader) Close() error {
	done := false
	loader.stopOnce.Do(func() {
		loader.pipeline.Close()
		loader.cancel()
		done = true
	})
	if !done {
		return fmt.Errorf("already closed")
	}
	return loader.g.Wait()
}

func (loader *ClientDBLoader) SendInput(blockRange BlockRange, dbTx database.DBTx) {
	msg := LoaderMsg{
		blockRange,
		&dbTx,
		blockRange,
	}
	loader.pipeline.blockRangeCh <- msg
}

func (loader *ClientDBLoader) blockRangeHandler(src <-chan LoaderMsg, dst chan<- LoaderMsg) error {
	for msg := range src {
		blockRange := msg.data.(BlockRange)
		hashes, err := loader.client.GetBlockHashesByRange(blockRange.startBlockHeight, blockRange.endBlockHeight)
		if err != nil {
			return err
		}
		newMsg := LoaderMsg{msg.blockRange, msg.dbTx, hashes}
		dst <- newMsg
	}
	return nil
}

func (loader *ClientDBLoader) blockHashesHandler(src <-chan LoaderMsg, headerDst chan<- LoaderMsg, txDst chan<- LoaderMsg) error {
	for msg := range src {
		hashes := msg.data.([]*chainhash.Hash)
		blocks, err := loader.client.GetBlocks(hashes)
		if err != nil {
			return err
		}
		newHeadersMsg := LoaderMsg{msg.blockRange, msg.dbTx, blocks}
		headerDst <- newHeadersMsg
		newTxsMsg := LoaderMsg{msg.blockRange, msg.dbTx, blocks}
		txDst <- newTxsMsg
	}
	return nil
}

func (loader *ClientDBLoader) blocksToHeadersHandler(src <-chan LoaderMsg, dst chan<- LoaderMsg) error {
	for msg := range src {
		blocks := msg.data.([]*types.Block)
		headers, err := BlocksToHeaders(blocks)
		if err != nil {
			return err
		}
		newMsg := LoaderMsg{msg.blockRange, msg.dbTx, headers}
		dst <- newMsg

	}
	return nil
}

func (loader *ClientDBLoader) blocksToTxsHandler(src <-chan LoaderMsg, dst chan<- LoaderMsg) error {
	for msg := range src {
		blocks := msg.data.([]*types.Block)
		txs, err := BlocksToTxs(blocks)
		if err != nil {
			return err
		}
		newMsg := LoaderMsg{msg.blockRange, msg.dbTx, txs}
		dst <- newMsg
	}
	return nil
}

func (loader *ClientDBLoader) headersHandler(src <-chan LoaderMsg) error {
	for msg := range src {
		txs := msg.data.([]*types.BlockHeader)
		dbTx := *msg.dbTx
		for _, header := range txs {
			loader.dbTxMu.Lock()
			err := dbTx.AddBlockHeader(header)
			loader.dbTxMu.Unlock()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (loader *ClientDBLoader) txsHandler(src <-chan LoaderMsg) error {
	for msg := range src {
		txs := msg.data.([]*types.Transaction)
		dbTx := *msg.dbTx
		for _, tx := range txs {
			loader.dbTxMu.Lock()
			err := dbTx.AddTransaction(tx)
			loader.dbTxMu.Unlock()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func BlocksToHeaders(blocks []*types.Block) ([]*types.BlockHeader, error) {
	headers := make([]*types.BlockHeader, 0)
	for _, block := range blocks {
		headers = append(headers, &block.BlockHeader)
	}
	return headers, nil
}

func BlocksToTxs(blocks []*types.Block) ([]*types.Transaction, error) {
	txs := make([]*types.Transaction, 0)
	for _, block := range blocks {
		txs = append(txs, block.Transactions()...)
	}
	return txs, nil
}
