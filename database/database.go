package database

import (
	"context"

	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Database represents a database connection.
type Database interface {
	LastBlockhash() (*chainhash.Hash, error)
	NewDBTx() (DBTx, error)
	Close() error
}

type DBTx interface {
	AddBlockHeader(blockheaders *types.BlockHeader) error
	AddTransaction(tx *types.Transaction) error
	Commit() error
	Rollback() error
}

type DBConstructor func(ctx context.Context, opts DBOptions) (Database, error)

type DBOptions map[string]interface{}
