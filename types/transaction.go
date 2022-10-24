package types

import "github.com/btcsuite/btcd/btcjson"

// Transaction represents a Bitcoin transaction.
//
// Wraps btcjson.TxRawResult.
type Transaction struct {
	data  btcjson.TxRawResult
	block *Block
}

// NewTransaction creates a new transaction from raw json result.
func NewTransaction(tx btcjson.TxRawResult, block *Block) *Transaction {
	return &Transaction{
		tx,
		block,
	}
}

// Hex returns the hex-encoded transaction.
func (tx *Transaction) Hex() string { return tx.data.Hex }

// TxID returns the transaction ID.
func (tx *Transaction) TxID() string { return tx.data.Txid }

// Hash returns the transaction hash (differs from txid for witness transactions).
func (tx *Transaction) Hash() string { return tx.data.Hash }

// Size returns the transaction size in bytes.
func (tx *Transaction) Size() int32 { return tx.data.Size }

// VSize returns the virtual transaction size (differs from size for witness transactions).
func (tx *Transaction) VSize() int32 { return tx.data.Vsize }

// Weight returns the transaction's weight (between vsize*4 - 3 and vsize*4).
func (tx *Transaction) Weight() int32 { return tx.data.Weight }

// Version returns the transaction version.
func (tx *Transaction) Version() uint32 { return tx.data.Version }

// LockTime returns the transaction lock time.
func (tx *Transaction) LockTime() uint32 { return tx.data.LockTime }

// Vin returns the transaction's inputs.
func (tx *Transaction) Vin() []btcjson.Vin { return tx.data.Vin }

// Vout returns the transaction's outputs.
func (tx *Transaction) Vout() []btcjson.Vout { return tx.data.Vout }

// BlockHash returns the hash of the block containing the transaction.
func (tx *Transaction) BlockHash() string { return tx.data.BlockHash }

// Confirmations returns the number of confirmations for the transaction.
func (tx *Transaction) Confirmations() uint64 { return tx.data.Confirmations }

// Time returns the time of the transaction.
func (tx *Transaction) Time() int64 { return tx.data.Time }

// BlockTime returns the time of the block containing the transaction.
func (tx *Transaction) BlockTime() int64 { return tx.data.Blocktime }

// Block returns the block containing the transaction.
func (tx *Transaction) Block() *Block {
	if tx.block == nil {
		return nil
	}
	cpy := *tx.block
	return &cpy
}

// String implements the Stringer interface.
func (tx *Transaction) String() string {
	return tx.data.Txid
}
