package client

import (
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Client represents a connection to a bitcoin node.
type Client interface {
	// GetBlockHashesByRange returns block hashes from the server given a range (inclusive) of block numbers.
	// Hashes are returned in order from `minBlockNumber` to `maxBlockNumber`
	GetBlockHashesByRange(minBlockNumber, maxBlockNumber int64) ([]*chainhash.Hash, error)
	// GetBlockHeadersByRange returns block headers from the server given a list/range of block hashes.
	GetBlockHeaders(hashes []*chainhash.Hash) ([]*types.BlockHeader, error)
	// GetBlocksByRange returns blocks with transactions from the server given a list/range of block hashes.
	GetBlocks(hashes []*chainhash.Hash) ([]*types.Block, error)
	// Close closes the connection to the server.
	Close() error
}
