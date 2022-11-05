package client

import (
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// Client represents a connection to a bitcoin node.
type Client interface {
	// GetBestBlockHash returns the hash of the best block in the longest block
	// chain.
	GetBestBlockHash() (*chainhash.Hash, error)
	// GetBlockCount returns the number of blocks in the longest block chain.
	GetBlockCount() (int64, error)
	// GetBlockHash returns the hash of the block in the best block chain at the
	// given height.
	GetBlockHash(blockHeight int64) (*chainhash.Hash, error)
	// GetBlockHeader returns the blockheader from the server given its hash.
	GetBlockHeader(blockHash *chainhash.Hash) (*wire.BlockHeader, error)
	// GetBlock returns a raw block from the server given its hash.
	GetBlock(blockHash *chainhash.Hash) (*wire.MsgBlock, error)
	// GetBlockVerbose returns a data structure from the server with information
	// about a block given its hash.
	GetBlockVerbose(blockHash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error)
	// GetBlockVerboseTx returns a data structure from the server with information
	// about a block and its transactions given its hash.
	GetBlockVerboseTx(blockHash *chainhash.Hash) (*btcjson.GetBlockVerboseTxResult, error)
	// GetTxOut returns the transaction output info if it's unspent and
	// nil, otherwise.
	GetTxOut(txHash *chainhash.Hash, index uint32, mempool bool) (*btcjson.GetTxOutResult, error)
	// GetRawMempool returns the hashes of all transactions in the memory pool.
	GetRawMempool() ([]*chainhash.Hash, error)
	// GetBlockHashesByRange returns block hashes from the server given a range (inclusive) of block numbers.
	// Hashes are returned in order from `minBlockNumber` to `maxBlockNumber`
	GetBlockHashesByRange(minBlockNumber, maxBlockNumber int64) ([]*chainhash.Hash, error)
	// GetBlockHeadersByRange returns block headers from the server given a list/range of block hashes.
	GetBlockHeadersByRange(hashes []*chainhash.Hash) ([]*types.BlockHeader, error)
	// GetBlocksByRange returns blocks with transactions from the server given a list/range of block hashes.
	GetBlocksByRange(hashes []*chainhash.Hash) ([]*types.Block, error)
}
