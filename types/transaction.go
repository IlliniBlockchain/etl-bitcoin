package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
)

// Transaction represents a Bitcoin transaction.
//
// Wraps btcjson.TxRawResult.
type Transaction struct {
	data  btcjson.TxRawResult
	block *Block
	vin   []*Vin
	vout  []*Vout
}

// NewTransaction creates a new transaction from raw json result.
func NewTransaction(data btcjson.TxRawResult) *Transaction {
	tx := &Transaction{
		data: data,
		vin:  make([]*Vin, len(data.Vin)),
		vout: make([]*Vout, len(data.Vout)),
	}
	for i, vin := range data.Vin {
		v := vin
		tx.vin[i] = NewVin(&v)
	}
	for i, vout := range data.Vout {
		v := vout
		tx.vout[i] = NewVout(&v)
	}
	return tx
}

// WithBlock stores a reference to the block that contains the transaction.
func (tx *Transaction) WithBlock(block *Block) (*Transaction, error) {
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}
	if tx.BlockHash() != block.Hash() {
		return nil, fmt.Errorf("block hash mismatch: %s != %s", tx.BlockHash(), block.Hash())
	}
	cpy := *block
	tx.block = &cpy
	return tx, nil
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
func (tx *Transaction) Vin() []*Vin { return tx.vin }

// Vout returns the transaction's outputs.
func (tx *Transaction) Vout() []*Vout { return tx.vout }

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

// Vin represents a Bitcoin transaction input.
type Vin struct {
	data *btcjson.Vin
}

// NewVin creates a new Vin from raw json result.
func NewVin(vin *btcjson.Vin) *Vin {
	return &Vin{
		data: vin,
	}
}

// Coinbase returns the coinbase data.
func (v *Vin) Coinbase() string { return v.data.Coinbase }

// TxID returns the transaction ID.
func (v *Vin) TxID() string { return v.data.Txid }

// Vout returns the output index.
func (v *Vin) Vout() uint32 { return v.data.Vout }

// Sequence returns the sequence number.
func (v *Vin) Sequence() uint32 { return v.data.Sequence }

// ScriptSig returns the scriptSig.
func (v *Vin) ScriptSig() *btcjson.ScriptSig {
	if v.data.ScriptSig == nil {
		return nil
	}
	cpy := *v.data.ScriptSig
	return &cpy
}

// Witness returns the witnesses of the input.
func (v *Vin) Witness() []string {
	if v.data.Witness == nil {
		return nil
	}
	cpy := make([]string, len(v.data.Witness))
	copy(cpy, v.data.Witness)
	return cpy
}

// IsCoinBase returns a bool to show if a Vin is a Coinbase one or not.
func (v *Vin) IsCoinbase() bool {
	return len(v.data.Coinbase) > 0
}

// Vout represents a Bitcoin transaction output.
type Vout struct {
	data *btcjson.Vout
}

// NewVout returns a new instance of a transaction output.
func NewVout(vout *btcjson.Vout) *Vout {
	return &Vout{
		data: vout,
	}
}

// Value returns the value of the output.
func (v *Vout) Value() float64 { return v.data.Value }

// N returns the index of the output.
func (v *Vout) N() uint32 { return v.data.N }

// ScriptPubKey returns the script of the output.
func (v *Vout) ScriptPubKey() btcjson.ScriptPubKeyResult { return v.data.ScriptPubKey }
