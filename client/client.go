package client

import (
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
	// GetBlocksByRange returns raw blocks from the server given a range of block numbers.
	GetBlocksByRange(minBlockNumber, maxBlockNumber int64) ([]*wire.MsgBlock, error)
	// GetBlocksVerboseByRange returns data structures from the server with information
	// about block given a range of block numbers.
	GetBlocksVerboseByRange(minBlockNumber, maxBlockNumber int64) ([]*btcjson.GetBlockVerboseResult, error)
	// GetBlocksVerboseTxByRange returns data structures from the server with information
	// about blocks and their transactions given a range of block numbers.
	GetBlocksVerboseTxByRange(minBlockNumber, maxBlockNumber int64) ([]*btcjson.GetBlockVerboseTxResult, error)
}
