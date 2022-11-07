package loader

import (
	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Loader represents loading process from a bitcoin client.
type XLoader interface {
}

// Approach 1: Loaders just implement functions related to loading a certain thing
type IBlockHashLoader interface {
}

type BlockHashLoader struct {
}

type IBlockLoader interface {
}

type BlockLoader struct {
}

type ITxLoader interface {
}

type TxLoader struct {
}

// Approach 2: Just write some functions that are needed and reorganize later
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

func BlocksToDisk(db database.Database, blocks []*types.Block) {

}

func TxsToDisk(db database.Database, txs []*types.Transaction) {

}

// BlocksFromDiskToDB? TxsFromDiskToDB?
