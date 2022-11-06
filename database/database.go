package database

import "github.com/IlliniBlockchain/etl-bitcoin/types"

// Database represents a database connection.
type Database interface {
	// Ideating the main processes
	SaveBlocks()
	UploadBlocks()
	SaveTransactions()
	UploadTransactions()
}

type Neo4j struct {
}

// Implement Database interface
func (db *Neo4j) SaveBlocks() {

}

func (db *Neo4j) UploadBlocks() {

}

func (db *Neo4j) SaveTransactions() {

}

func (db *Neo4j) UploadTransactions() {

}

// Other helper functions for Neo4j
func (db *Neo4j) BlockToRecord(*types.Block) []string {
	return nil
}

func (db *Neo4j) BlockHeaderToRecord(*types.BlockHeader) []string {
	return nil
}

func (db *Neo4j) TxToRecord(*types.Transaction) []string {
	return nil
}
