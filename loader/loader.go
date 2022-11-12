package loader

import (
	"github.com/IlliniBlockchain/etl-bitcoin/client"
	rpcclient "github.com/IlliniBlockchain/etl-bitcoin/client/rpc"
	"github.com/IlliniBlockchain/etl-bitcoin/csv"
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
type ClientDatabaseLoader interface {
	Run()
}

type ClientDBLoaderConstructor func(client client.Client, db database.Database, opts LoaderOptions) (ClientDatabaseLoader, error)

// DatabaseLoader represents any process of loading data from one type of
// Database to another.
type DatabaseLoader interface {
}

type DBLoaderConstructor func(srcDb, dstDb database.Database, opts LoaderOptions) (DBLoaderConstructor, error)

// RPCCSVLoader represents the loading process for retrieving data from
// a full node and saving it to disk as CSVs. Will implement
// ClientDatabaseLoader.
type RPCCSVLoader struct {
	client *rpcclient.RPCClient
	db     *csv.CSVDatabase
}

// CSVNeo4jLoader represents the loading process for uploading CSV data
// into Neo4j. Will implement DatabaseLoader.
type CSVNeo4jLoader struct {
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
