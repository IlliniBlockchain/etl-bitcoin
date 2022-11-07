package database

import (
	"fmt"

	"github.com/IlliniBlockchain/etl-bitcoin/types"
)

// Database represents a database connection.
type Database interface {
	// Ideating the main processes
	SaveBlocks([]*types.Block)
	UploadBlocks()
	SaveTransactions([]*types.Transaction)
	UploadTransactions()
}

type Neo4j struct {
}

func NewNeo4j() *Neo4j {
	db := Neo4j{}
	return &db
}

// Implement Database interface
func (db *Neo4j) SaveBlocks(blocks []*types.Block) {
	for _, block := range blocks {
		fmt.Println(block.Hash())
	}

}

func (db *Neo4j) UploadBlocks() {

}

func (db *Neo4j) SaveTransactions(txs []*types.Transaction) {
	for _, tx := range txs {
		fmt.Println(tx.Hash())
	}

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
