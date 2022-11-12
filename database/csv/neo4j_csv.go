package csv

import (
	"context"
	"runtime"

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

type Neo4jCSVDatabase struct {
	*CSVDatabase
}

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

func (db *Neo4jCSVDatabase) LastBlockhash() (*chainhash.Hash, error) {
	return nil, nil
}

func (db *Neo4jCSVDatabase) NewDBTx() (database.DBTx, error) {
	dbTx := &Neo4jCSVDBTx{db: db, data: make(map[string][]CSVRecord)}
	for fileKey := range defaultNeo4jFilePaths {
		dbTx.data[fileKey] = make([]CSVRecord, 0)
	}
	return dbTx, nil
}

type Neo4jCSVDBTx struct {
	db *Neo4jCSVDatabase

	data map[string][]CSVRecord
}

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

func (dbTx Neo4jCSVDBTx) AddBlockHeader(bh *types.BlockHeader) {
	dbTx.data[blockKey] = append(dbTx.data[blockKey], csvBlockHeader{bh})
}

func (dbTx Neo4jCSVDBTx) AddTransaction(tx *types.Transaction) {
	dbTx.data[txKey] = append(dbTx.data[txKey], csvTransaction{tx})
}

type csvBlockHeader struct {
	*types.BlockHeader
}

func (bh csvBlockHeader) Headers() []string {
	return []string{"block_id", "version", "merkle_root", "time", "nonce"}
}

func (bh csvBlockHeader) Row() []any {
	return []any{bh.Hash(), bh.Version(), bh.MerkleRoot(), bh.Time(), bh.Nonce()}
}

type csvTransaction struct {
	*types.Transaction
}

func (tx csvTransaction) Headers() []string {
	return []string{"tx_id", "block_id", "version", "locktime"}
}

func (tx csvTransaction) Row() []any {
	return []any{tx.TxID(), tx.Block().Hash(), tx.Version(), tx.LockTime()}
}
