package loader

import (
	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// General idea for loader:
// Represents a single entire process of getting data from one place to another.
// It takes in clients (RPC or DB) and conducts a designated process by using them.

// ClientDatabaseLoader represents any process of loading data from a
// Client to a Database
type ClientDatabaseLoader interface {
}

// DatabaseLoader represents any process of loading data from one type of
// Database to another.
type DatabaseLoader interface {
}

// RPCCSVLoader represents the loading process for retrieving data from
// a full node and saving it to disk as CSVs. Will implement
// ClientDatabaseLoader.
type RPCCSVLoader struct {
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
