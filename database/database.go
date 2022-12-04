package database

import (
	"context"

	"github.com/IlliniBlockchain/etl-bitcoin/types"
)

// Database represents a database connection.
type Database interface {
	// LastBlockNumber returns the block height of the last block committed to the database.
	LastBlockNumber() (int64, error)
	// NewDBTx returns a new DBTx.
	NewDBTx() (DBTx, error)
	// Close closes the database connection. Only callable once.
	Close() error
}

// DBTx represents a database transaction to make atomic changes to a database.
type DBTx interface {
	// AddBlockHeader adds block header data to this database transaction.
	AddBlockHeader(blockheaders *types.BlockHeader)
	// AddTransaction adds transaction data to this database transaction.
	AddTransaction(tx *types.Transaction)
	// Commit commits this database transaction to a database.
	Commit() error
}

// DBConstructor represents a function that constructs a database connection.
type DBConstructor func(ctx context.Context, opts DBOptions) (Database, error)

// DBOptions represents a map of arbitrary options when constructing a database connection.
type DBOptions map[string]interface{}

// GetOpt returns the value of the option with the given key. If the option is not set, the default value is returned.
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
