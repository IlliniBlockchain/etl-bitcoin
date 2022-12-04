package loader

import (
	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// blockRangeHandler is a LoaderFunc that uses a block range to
// retrieve a list of block hashes.
func blockRangeHandler(client client.Client, blockRange BlockRange) ([]*chainhash.Hash, error) {
	return client.GetBlockHashesByRange(blockRange.Start, blockRange.End)
}

// blockHashHandler is a LoaderFunc that uses a list of block
// hashes to retrieve block header and transaction data.
func blockHashHandler(client client.Client, hashes []*chainhash.Hash) ([]*types.Block, error) {
	return client.GetBlocks(hashes)
}

// blockHandler is a LoaderSinkFunc that fills a database transaction with
// block headers and transactions.
func blockHandler(dbTx database.DBTx, blocks []*types.Block) error {
	for _, block := range blocks {
		dbTx.AddBlockHeader(block.BlockHeader)
		for _, tx := range block.Transactions() {
			dbTx.AddTransaction(tx)
		}
	}
	return nil
}
