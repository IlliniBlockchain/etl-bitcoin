package loader

import (
	"fmt"

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
	Run(BlockRange, database.DBTx)
	Close()
}

type ClientDBLoaderConstructor func(client client.Client, db database.Database, opts LoaderOptions) (Loader, error)

type ClientDBLoader struct {
	client     client.Client
	db         database.Database
	maxWorkers int
	msgs       chan LoaderMsg
}

type LoaderMsg struct {
	t          LoaderMsgType
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

func NewClientDBLoader(client client.Client, db database.Database, opts LoaderOptions, maxWorkers int) (*ClientDBLoader, error) {
	// initialize struct
	msgs := make(chan LoaderMsg)
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
	msg := LoaderMsg{
		Range,
		blockRange,
		&dbTx,
		blockRange,
	}
	loader.msgs <- msg
}

func (loader *ClientDBLoader) loaderWorker() error {
	for msg := range loader.msgs {
		switch msg.t {
		case Range:
			blockRange := msg.data.(BlockRange)
			hashes, err := loader.client.GetBlockHashesByRange(blockRange.startBlockHeight, blockRange.endBlockHeight)
			if err != nil {
				return err
			}
			newMsg := LoaderMsg{Hashes, msg.blockRange, msg.dbTx, hashes}
			loader.msgs <- newMsg

		case Hashes:
			hashes := msg.data.([]*chainhash.Hash)
			blocks, err := loader.client.GetBlocks(hashes)
			if err != nil {
				return err
			}
			newHeadersMsg := LoaderMsg{BlocksForHeaders, msg.blockRange, msg.dbTx, blocks}
			loader.msgs <- newHeadersMsg
			newTxsMsg := LoaderMsg{BlocksForTxs, msg.blockRange, msg.dbTx, blocks}
			loader.msgs <- newTxsMsg

		case BlocksForHeaders:
			blocks := msg.data.([]*types.Block)
			headers, err := BlocksToHeaders(blocks)
			if err != nil {
				return err
			}
			newMsg := LoaderMsg{Headers, msg.blockRange, msg.dbTx, headers}
			loader.msgs <- newMsg

		case BlocksForTxs:
			blocks := msg.data.([]*types.Block)
			txs, err := BlocksToTxs(blocks)
			if err != nil {
				return err
			}
			newMsg := LoaderMsg{Txs, msg.blockRange, msg.dbTx, txs}
			loader.msgs <- newMsg

		case Headers:
			headers := msg.data.([]*types.BlockHeader)
			dbTx := nil // will retrieve from somewhere
			for _, header := range headers {
				dbTx.AddBlockHeader(header)
			}

		case Txs:
			txs := msg.data.([]*types.Transaction)
			dbTx := nil
			for _, tx := range txs {
				dbTx.AddTransaction(tx)
			}

		default:
			return fmt.Errorf("Unknown LoaderMsgType %d", msg.t)
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
