package csv

import (
	"context"
	"runtime"
	"strconv"

	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

var (
	blockKey   = "blocks"
	txKey      = "txs"
	outputKey  = "outputs"
	addressKey = "addresses"
)

var defaultNeo4jFilePaths = map[string]string{blockKey: "blocks.csv", txKey: "transactions.csv", outputKey: "outputs.csv", addressKey: "addresses.csv"}

// Neo4jCSVDatabase is a database.Database implementation that writes to CSV files formatted for Neo4j.
type Neo4jCSVDatabase struct {
	*CSVDatabase
}

// NewNeo4jCSVDatabase creates a new Neo4jCSVDatabase.
//
// Implements database.DBConstructor.
func NewNeo4jCSVDatabase(ctx context.Context, opts database.DBOptions) (database.Database, error) {
	filePaths := make(map[string]string)
	for fileKey, defaultFilePath := range defaultNeo4jFilePaths {
		filePath, err := database.GetOpt(opts, fileKey, defaultFilePath)
		if err != nil {
			return nil, err
		}
		filePaths[fileKey] = filePath
	}
	maxWorkers, err := database.GetOpt(opts, "maxWorkers", runtime.NumCPU())
	if err != nil {
		return nil, err
	}
	csvDB, err := NewCSVDatabase(ctx, filePaths, maxWorkers)
	if err != nil {
		return nil, err
	}
	return &Neo4jCSVDatabase{csvDB}, nil
}

// LastBlockHash returns the hash of the last block in the database.
//
// Implements database.Database.
func (db *Neo4jCSVDatabase) LastBlockhash() (*chainhash.Hash, error) {
	return nil, nil
}

// NewDBTx creates a new database transaction.
//
// Implements database.Database.
func (db *Neo4jCSVDatabase) NewDBTx() (database.DBTx, error) {
	dbTx := &Neo4jCSVDBTx{db: db, data: make(map[string][]CSVRecord)}
	for fileKey := range defaultNeo4jFilePaths {
		dbTx.data[fileKey] = make([]CSVRecord, 0)
	}
	return dbTx, nil
}

// Neo4jCSVDBTx is a database.DBTx implementation that processes a batch of operations to a `Neo4jCSVDatabase`.
type Neo4jCSVDBTx struct {
	db *Neo4jCSVDatabase

	data map[string][]CSVRecord
}

// Commit commits the transaction.
//
// Implements database.DBTx.
func (dbTx Neo4jCSVDBTx) Commit() error {
	msgs := make([]CSVMsg, 0)
	for fileKey, records := range dbTx.data {
		if len(records) == 0 {
			continue
		}
		msgs = append(msgs, NewCSVInsertMsg(fileKey, records))
	}
	return dbTx.db.SendMsgs(msgs)
}

// AddBlockHeader processes a block header for a database.
//
// Implements database.DBTx.
func (dbTx Neo4jCSVDBTx) AddBlockHeader(bh *types.BlockHeader) {
	dbTx.data[blockKey] = append(dbTx.data[blockKey], csvBlockHeader{bh})
}

// AddTransaction processes a transaction for a database.
//
// Implements database.DBTx.
func (dbTx Neo4jCSVDBTx) AddTransaction(tx *types.Transaction) {
	dbTx.data[txKey] = append(dbTx.data[txKey], csvTransaction{tx})
}

type csvBlockHeader struct {
	*types.BlockHeader
}

// Headers returns the headers for the CSV file.
//
// Implements CSVRecord.
func (bh csvBlockHeader) Headers() []string {
	return []string{
		"blockID:id",
		"height:int",
		"time:int",
		"size:int",
		"difficulty:double",
		"nonce:int",
	}
}

// Rows returns the row for the CSV file.
//
// Implements CSVRecord.
func (bh csvBlockHeader) Row() []string {
	return []string{
		bh.Hash(),
		strconv.FormatInt(bh.Height(), 10),
		strconv.FormatInt(bh.Time(), 10),
		strconv.FormatInt(int64(bh.Size()), 10),
		strconv.FormatFloat(bh.Difficulty(), 'f', -1, 64),
		strconv.FormatUint(uint64(bh.Nonce()), 10),
	}
}

type csvTransaction struct {
	*types.Transaction
}

// Headers returns the headers for the CSV file.
//
// Implements CSVRecord.
func (tx csvTransaction) Headers() []string {
	return []string{
		"txID:id",
		"size:int",
		"time:int",
		"lockTime:int",
	}
}

// Rows returns the row for the CSV file.
//
// Implements CSVRecord.
func (tx csvTransaction) Row() []string {
	return []string{
		tx.TxID(),
		strconv.FormatInt(int64(tx.Size()), 10),
		strconv.FormatInt(int64(tx.Time()), 10),
		strconv.FormatUint(uint64(tx.LockTime()), 10),
	}
}
