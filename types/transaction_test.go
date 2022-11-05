package types

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/stretchr/testify/suite"
)

type TransactionTestSuite struct {
	suite.Suite
	tx btcjson.TxRawResult
}

func (s *TransactionTestSuite) SetupTest() {
	// Read the raw transaction from file.
	data, err := os.ReadFile("testdata/transaction.json")
	s.NoError(err, "failed to read transaction test data")
	s.NoError(json.Unmarshal(data, &s.tx), "failed to parse transaction test data")
}

func (s *TransactionTestSuite) TestTransactionData() {
	tx := NewTransaction(s.tx)
	s.Equal(s.tx.Hex, tx.Hex())
	s.Equal(s.tx.Txid, tx.TxID())
	s.Equal(s.tx.Hash, tx.Hash())
	s.Equal(s.tx.Size, tx.Size())
	s.Equal(s.tx.Vsize, tx.VSize())
	s.Equal(s.tx.Weight, tx.Weight())
	s.Equal(s.tx.Version, tx.Version())
	s.Equal(s.tx.LockTime, tx.LockTime())
	s.Equal(s.tx.Vin, tx.Vin())
	s.Equal(s.tx.Vout, tx.Vout())
	s.Equal(s.tx.BlockHash, tx.BlockHash())
	s.Equal(s.tx.Confirmations, tx.Confirmations())
	s.Equal(s.tx.Time, tx.Time())
	s.Equal(s.tx.Blocktime, tx.BlockTime())
	s.Nil(tx.Block())
}

func (s *TransactionTestSuite) TestTransactionWithBlock() {
	/// Block 409008 contains the transaction.
	blk := &Block{
		BlockHeader: BlockHeader{
			data: btcjson.GetBlockVerboseResult{
				Hash: "0000000000000000042450ad2be4f2b6439ed39f70716a7575440d462cf165d9",
			},
		},
	}
	tx, err := NewTransactionWithBlock(s.tx, blk)
	s.NoError(err, "failed to create transaction with block")
	s.Equal(*blk, *tx.Block(), "block should be unchanged")
	s.NotSame(blk, tx.Block(), "block should be a copy")

	/// Block 1000 does not contain the transaction.
	blk = &Block{
		BlockHeader: BlockHeader{
			data: btcjson.GetBlockVerboseResult{
				Hash: "00000000c937983704a73af28acdec37b049d214adbda81d7e2a3dd146f6ed09",
			},
		},
	}
	tx, err = NewTransactionWithBlock(s.tx, blk)
	s.Error(err, "block does not contain transaction")
	s.Nil(tx, "tx should be nil")
}

func TestTransactionTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionTestSuite))
}
