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

func GetOpt[T any](opts DBOptions, key string, def T) (T, error) {
	opt, ok := opts[key]
	if !ok {
		return def, nil
	}
	val, ok := opt.(T)
	if !ok {
		return def, ErrInvalidOptionType{opt, def}
	}
	return val, nil
}
