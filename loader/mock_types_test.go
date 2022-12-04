package loader

import (
	"fmt"

	"github.com/IlliniBlockchain/etl-bitcoin/database"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type MockClient struct {
	blocks         []*types.Block
	maxBlockNumber int64
	minBlockNumber int64
}

func NewMockClient(blocks []*types.Block) *MockClient {
	return &MockClient{
		blocks:         blocks,
		maxBlockNumber: int64(len(blocks)) - 1,
		minBlockNumber: 0,
	}
}

func (c *MockClient) MaxBlockNumber() int64 {
	return c.maxBlockNumber
}

func (c *MockClient) MinBlockNumber() int64 {
	return c.minBlockNumber
}

func (c *MockClient) Blocks() []*types.Block {
	if c.blocks == nil {
		return nil
	}
	blocksCpy := make([]*types.Block, len(c.blocks))
	for i, block := range c.blocks {
		cpy := *block
		blocksCpy[i] = &cpy
	}
	return blocksCpy
}

func (c *MockClient) GetBlockHashesByRange(minBlockNumber, maxBlockNumber int64) ([]*chainhash.Hash, error) {
	// return error if minBlockNumber is less than minBlockNumber or maxBlockNumber is greater than maxBlockNumber
	if minBlockNumber < c.minBlockNumber || maxBlockNumber > c.maxBlockNumber {
		return nil, fmt.Errorf("invalid block range for mock client")
	}
	hashes := make([]*chainhash.Hash, 0)
	for _, block := range c.blocks {
		hash, err := chainhash.NewHashFromStr(block.BlockHeader.Hash())
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, nil
}

func (c *MockClient) GetBlockHeaders(hashes []*chainhash.Hash) ([]*types.BlockHeader, error) {
	// getblock headers by searching for the block
	headers := make([]*types.BlockHeader, 0)
	for _, hash := range hashes {
		found := false
		for _, block := range c.blocks {
			if block.BlockHeader.Hash() == hash.String() {
				headers = append(headers, block.BlockHeader)
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("block not found")
		}
	}
	return headers, nil
}

func (c *MockClient) GetBlocks(hashes []*chainhash.Hash) ([]*types.Block, error) {
	// get blocks by searching for the block
	blocks := make([]*types.Block, 0)
	for _, hash := range hashes {
		found := false
		for _, block := range c.blocks {
			if block.BlockHeader.Hash() == hash.String() {
				blocks = append(blocks, block)
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("block not found")
		}
	}
	return blocks, nil
}

type MockDatabase struct {
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

func (db *MockDatabase) LastBlockNumber() (int64, error) {
	return 0, nil
}

func (db *MockDatabase) NewDBTx() (database.DBTx, error) {
	return db.NewMockDBTx(), nil
}

func (db *MockDatabase) NewMockDBTx() *MockDBTx {
	return &MockDBTx{
		receivedBlockHeaders: make([]*types.BlockHeader, 0),
		receivedTxs:          make([]*types.Transaction, 0),
		committed:            false,
	}
}

func (db *MockDatabase) Close() error {
	return nil
}

type MockDBTx struct {
	receivedBlockHeaders []*types.BlockHeader
	receivedTxs          []*types.Transaction
	committed            bool
}

func (tx *MockDBTx) AddBlockHeader(header *types.BlockHeader) {
	tx.receivedBlockHeaders = append(tx.receivedBlockHeaders, header)
}

func (tx *MockDBTx) AddTransaction(txn *types.Transaction) {
	tx.receivedTxs = append(tx.receivedTxs, txn)
}

func (tx *MockDBTx) Commit() error {
	tx.committed = true
	return nil
}

func (tx *MockDBTx) ReceivedBlockHeaders() []*types.BlockHeader {
	headers := make([]*types.BlockHeader, len(tx.receivedBlockHeaders))
	for i, header := range tx.receivedBlockHeaders {
		cpy := *header
		headers[i] = &cpy
	}
	return headers
}

func (tx *MockDBTx) ReceivedTxs() []*types.Transaction {
	txs := make([]*types.Transaction, len(tx.receivedTxs))
	for i, tx := range tx.receivedTxs {
		cpy := *tx
		txs[i] = &cpy
	}
	return txs
}

func (tx *MockDBTx) Committed() bool {
	return tx.committed
}
