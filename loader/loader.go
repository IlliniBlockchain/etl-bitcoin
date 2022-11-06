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
func BlockRangesToHashes(client *client.Client, startBlockHeight, endBlockHeight int64) []*chainhash.Hash {
	return nil
}

func HashesToBlocks(client *client.Client, hashes []*chainhash.Hash) []*types.Block {
	return nil
}

func BlocksToTxs([]*types.Block) []*types.Transaction {
	return nil
}

func BlocksToDisk(db *database.Database, blocks []*types.Block) {

}

func TxsToDisk(db *database.Database, txs []*types.Transaction) {

}

// BlocksFromDiskToDB? TxsFromDiskToDB?
