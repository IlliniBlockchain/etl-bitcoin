package loader

import (
	"fmt"
	"sync"

	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
	Run()
}

type ClientDBLoaderConstructor func(client client.Client, db database.Database, opts LoaderOptions) (Loader, error)

type ClientDBLoader struct {
	client client.Client
	db     database.Database

	maxWorkers int
	dbTxs      sync.Map
	msgs       chan interface{}

	// rangeChan chan BlockRange
	// hashChan  chan []*chainhash.Hash
	// blockChan chan []*types.Block
}

type BlockRange struct {
	startBlockHeight int64
	endBlockHeight   int64
}

func NewClientDBLoader(client client.Client, db database.Database, opts LoaderOptions, maxWorkers int) (*ClientDBLoader, error) {
	// initialize channels
	msgs := make(chan interface{})
	loader := &ClientDBLoader{
		client:     client,
		db:         db,
		msgs:       msgs,
		maxWorkers: maxWorkers,
	}

	// start workers
	for i := 0; i < maxWorkers; i++ {
		go loader.loaderWorker()
	}

	return loader, nil
}

func (loader *ClientDBLoader) Run(blockRange BlockRange, dbTx database.DBTx) {
	// Associate blockRange with tx
	loader.dbTxs.Store(blockRange, &dbTx)
	// Send blockRange
	loader.msgs <- blockRange
}

func (loader *ClientDBLoader) loaderWorker() error {
	for msg := range loader.msgs {
		// split this into individual functions to handle
		switch msg := msg.(type) {
		case BlockRange:
			blockRange := msg
			hashes, err := BlockRangesToHashes(loader.client, blockRange.startBlockHeight, blockRange.endBlockHeight)
			if err != nil {
				return err
			}
			loader.msgs <- hashes
		case []*chainhash.Hash:
			hashes := msg
			blocks, err := HashesToBlocks(loader.client, hashes)
			if err != nil {
				return err
			}
			loader.msgs <- blocks
		case []*types.Block:
			blocks := msg
			headers := make([]*types.BlockHeader, 0, len(blocks))
			txs := make([]*types.Transaction, 0, len(blocks))
			for _, block := range blocks {
				headers = append(headers, &block.BlockHeader)
				txs = append(txs, block.Transactions()...)
			}
			loader.msgs <- headers
			loader.msgs <- txs
		case []*types.BlockHeader:
			headers := msg
			dbTx := nil // will retrieve from somewhere
			for _, header := range headers {
				dbTx.AddBlockHeader(header)
			}
		case []*types.Transaction:
			txs := msg
			dbTx := nil
			for _, tx := range txs {
				dbTx.AddTransaction(tx)
			}
		default:
			return fmt.Errorf("unknown message type %T", msg)
		}
	}
	return nil
}

// Just write some functions that are needed and reorganize later
func BlockRangesToHashes(client client.Client, startBlockHeight, endBlockHeight int64) ([]*chainhash.Hash, error) {
	hashes, err := client.GetBlockHashesByRange(startBlockHeight, endBlockHeight)
	if err != nil {
		return nil, err
	}
	return hashes, nil
}

func HashesToBlocks(client client.Client, hashes []*chainhash.Hash) ([]*types.Block, error) {
	blocks, err := client.GetBlocks(hashes)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func BlocksToTxs(blocks []*types.Block) ([]*types.Transaction, error) {
	txs := make([]*types.Transaction, 0)
	for _, block := range blocks {
		txs = append(txs, block.Transactions()...)
	}
	return txs, nil
}
