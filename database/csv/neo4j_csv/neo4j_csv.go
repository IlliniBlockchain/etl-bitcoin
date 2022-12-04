package neo4j_csv

import (
	"context"
	"runtime"
	"strconv"

	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"github.com/IlliniBlockchain/etl-bitcoin/database/csv"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
)

var (
	// Nodes
	BlockKey   = "blocks"
	TxKey      = "transactions"
	OutputKey  = "outputs"
	AddressKey = "addresses"

	// Relationships
	ChainKey    = "chain"    // Block --> Block
	CoinbaseKey = "coinbase" // Output --> Tx
	IncludeKey  = "include"  // Tx -----> Block
	InKey       = "in"       // Output -> Tx
	OutKey      = "out"      // Tx -----> Output
	LockedKey   = "locked"   // Output -> Address

	fileKeys = []string{BlockKey, TxKey, OutputKey, AddressKey, ChainKey, CoinbaseKey, IncludeKey, InKey, OutKey, LockedKey}
)

// Database is a database.Database implementation that writes to CSV files formatted for Neo4j.
type Database struct {
	*csv.CSVDatabase
}

// NewDatabase creates a new Neo4j CSV Database.
//
// Implements database.DBConstructor.
func NewDatabase(ctx context.Context, opts database.DBOptions) (database.Database, error) {
	filePaths := make(map[string]string)
	for _, fileKey := range fileKeys {
		filePath, err := database.GetOpt(opts, fileKey, fileKey+".csv")
		if err != nil {
			return nil, err
		}
		filePaths[fileKey] = filePath
	}
	maxWorkers, err := database.GetOpt(opts, "maxWorkers", runtime.NumCPU())
	if err != nil {
		return nil, err
	}
	csvDB, err := csv.NewCSVDatabase(ctx, filePaths, maxWorkers)
	if err != nil {
		return nil, err
	}
	return &Database{csvDB}, nil
}

// LastBlockNumber returns the height of the last block in the database.
//
// Implements database.Database.
func (db *Database) LastBlockNumber() (int64, error) {
	lastBlockRead := csv.NewCSVReadMsg(BlockKey, -1, 1)
	if err := db.SendMsg(lastBlockRead); err != nil {
		return 0, err
	}
	if len(lastBlockRead.Records) == 0 {
		return 0, nil
	}
	blockHeightStr, err := csv.GetRowField(csvBlockHeader{}.Headers(), lastBlockRead.Records[0], "height:int")
	if err != nil {
		return 0, err
	}
	blockHeight, err := strconv.ParseInt(blockHeightStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return blockHeight, nil
}

// NewDBTx creates a new database transaction.
//
// Implements database.Database.
func (db *Database) NewDBTx() (database.DBTx, error) {
	dbTx := &DBTx{db: db, data: make(map[string][]csv.CSVRecord), addresses: make(map[string]struct{})}
	for _, fileKey := range fileKeys {
		dbTx.data[fileKey] = make([]csv.CSVRecord, 0)
	}
	return dbTx, nil
}

// DBTx is a database.DBTx implementation that processes a batch of operations to a `Database`.
type DBTx struct {
	db *Database

	data      map[string][]csv.CSVRecord
	addresses map[string]struct{}
}

// Commit commits the transaction.
//
// Implements database.DBTx.
func (dbTx DBTx) Commit() error {
	// Add addresses
	for address := range dbTx.addresses {
		dbTx.data[AddressKey] = append(dbTx.data[AddressKey], csvAddress{address})
	}

	msgs := make([]csv.CSVMsg, 0)
	for fileKey, records := range dbTx.data {
		if len(records) == 0 {
			continue
		}
		msgs = append(msgs, csv.NewCSVInsertMsg(fileKey, records))
	}
	return dbTx.db.SendMsgs(msgs)
}

// AddBlockHeader processes a block header for a database.
//
// Implements database.DBTx.
func (dbTx DBTx) AddBlockHeader(bh *types.BlockHeader) {
	dbTx.data[BlockKey] = append(dbTx.data[BlockKey], csvBlockHeader{bh})
	if bh.Height() > 0 {
		// Genesis block has no parent
		dbTx.data[ChainKey] = append(dbTx.data[ChainKey], csvChainRelation{bh})
	}
	dbTx.data[OutputKey] = append(dbTx.data[OutputKey], csvCoinbaseVout{bh})
	dbTx.data[CoinbaseKey] = append(dbTx.data[CoinbaseKey], csvCoinbaseRelation{bh})
}

// AddTransaction processes a transaction for a database.
//
// Implements database.DBTx.
func (dbTx DBTx) AddTransaction(tx *types.Transaction) {
	dbTx.data[TxKey] = append(dbTx.data[TxKey], csvTransaction{tx})
	dbTx.data[IncludeKey] = append(dbTx.data[IncludeKey], csvIncludeRelation{tx})

	if tx.Vin()[0].IsCoinbase() {
		dbTx.data[InKey] = append(dbTx.data[InKey], csvCoinbaseInRelation{tx})
	} else {
		for _, input := range tx.Vin() {
			dbTx.data[InKey] = append(dbTx.data[InKey], csvInRelation{input, tx})
		}
	}

	for _, output := range tx.Vout() {
		dbTx.data[OutputKey] = append(dbTx.data[OutputKey], csvVout{output, tx})
		dbTx.data[OutKey] = append(dbTx.data[OutKey], csvOutRelation{output, tx})
		for _, address := range output.ScriptPubKey().Addresses {
			dbTx.addresses[address] = struct{}{}
			dbTx.data[LockedKey] = append(dbTx.data[LockedKey], csvLockedRelation{output, tx, address})
		}
	}
}
