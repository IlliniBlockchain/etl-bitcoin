package types

import "github.com/btcsuite/btcd/btcjson"

// BlockHeader represents a Bitcoin block header.
//
// Wraps btcjson.GetBlockVerboseResult.
type BlockHeader struct {
	data btcjson.GetBlockVerboseResult
}

// NewBlockHeader returns a new instance of a block header from a raw json block header.
func NewBlockHeader(blockHeader btcjson.GetBlockHeaderVerboseResult) *BlockHeader {
	return &BlockHeader{
		btcjson.GetBlockVerboseResult{
			Hash:          blockHeader.Hash,
			Confirmations: blockHeader.Confirmations,
			Height:        int64(blockHeader.Height),
			Version:       blockHeader.Version,
			VersionHex:    blockHeader.VersionHex,
			MerkleRoot:    blockHeader.MerkleRoot,
			Time:          blockHeader.Time,
			Nonce:         uint32(blockHeader.Nonce),
			Bits:          blockHeader.Bits,
			Difficulty:    blockHeader.Difficulty,
			PreviousHash:  blockHeader.PreviousHash,
			NextHash:      blockHeader.NextHash,
		},
	}
}

// NewBlockHeaderFromVerboseTx returns a new instance of a block header from a raw json block.
func NewBlockHeaderFromVerboseTx(blockHeader btcjson.GetBlockVerboseTxResult) *BlockHeader {
	return &BlockHeader{
		btcjson.GetBlockVerboseResult{
			Hash:          blockHeader.Hash,
			Confirmations: blockHeader.Confirmations,
			StrippedSize:  blockHeader.StrippedSize,
			Size:          blockHeader.Size,
			Weight:        blockHeader.Weight,
			Height:        blockHeader.Height,
			Version:       blockHeader.Version,
			VersionHex:    blockHeader.VersionHex,
			MerkleRoot:    blockHeader.MerkleRoot,
			Time:          blockHeader.Time,
			Nonce:         blockHeader.Nonce,
			Bits:          blockHeader.Bits,
			Difficulty:    blockHeader.Difficulty,
			PreviousHash:  blockHeader.PreviousHash,
			NextHash:      blockHeader.NextHash,
		},
	}
}

// Hash returns the hash of the block.
func (bh *BlockHeader) Hash() string { return bh.data.Hash }

// Confirmations returns the number of confirmations of the block.
func (bh *BlockHeader) Confirmations() int64 { return bh.data.Confirmations }

// StrippedSize returns the size of the block without the witness data.
func (bh *BlockHeader) StrippedSize() int32 { return bh.data.StrippedSize }

// Size returns the size of the block.
func (bh *BlockHeader) Size() int32 { return bh.data.Size }

// Weight returns the weight of the block as defined in BIP 141.
func (bh *BlockHeader) Weight() int32 { return bh.data.Weight }

// Height returns the block height or index.
func (bh *BlockHeader) Height() int64 { return bh.data.Height }

// Version returns the block version.
func (bh *BlockHeader) Version() int32 { return bh.data.Version }

// VersionHex returns the block version formatted in hexadecimal.
func (bh *BlockHeader) VersionHex() string { return bh.data.VersionHex }

// MerkleRoot returns the merkle root of the block.
func (bh *BlockHeader) MerkleRoot() string { return bh.data.MerkleRoot }

// Time returns the block time expressed in UNIX epoch time
func (bh *BlockHeader) Time() int64 { return bh.data.Time }

// Nonce returns the block nonce.
func (bh *BlockHeader) Nonce() uint32 { return bh.data.Nonce }

// Bits returns the bits of the block.
func (bh *BlockHeader) Bits() string { return bh.data.Bits }

// Difficulty returns the difficulty of the block.
func (bh *BlockHeader) Difficulty() float64 { return bh.data.Difficulty }

// PreviousHash returns the previous block hash.
func (bh *BlockHeader) PreviousHash() string { return bh.data.PreviousHash }

// NextHash returns the next block hash if it exists, else returns empty string.
func (bh *BlockHeader) NextHash() string { return bh.data.NextHash }

// String implements the Stringer interface.
func (bh *BlockHeader) String() string {
	return bh.data.Hash
}

// WithTransactions returns a Block with the given transactions.
func (bh *BlockHeader) WithTransactions(txs []*Transaction) *Block {
	return &Block{
		*bh,
		txs,
	}
}

// Block represents a block header with its transactions.
type Block struct {
	BlockHeader
	txs []*Transaction
}

// NewBlock returns a new instance of a block from a raw json block.
func NewBlock(blockVerbose btcjson.GetBlockVerboseTxResult) *Block {
	block := &Block{
		*NewBlockHeaderFromVerboseTx(blockVerbose),
		make([]*Transaction, len(blockVerbose.Tx)),
	}
	for i, tx := range blockVerbose.Tx {
		block.txs[i] = NewTransaction(tx, block)
	}
	return block
}

// Transactions returns the transactions of the block.
func (b *Block) Transactions() []*Transaction {
	if b.txs == nil {
		return nil
	}
	txsCpy := make([]*Transaction, len(b.txs))
	for i, tx := range b.txs {
		cpy := *tx
		txsCpy[i] = &cpy
	}
	return txsCpy
}
