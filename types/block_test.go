package types

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/stretchr/testify/suite"
)

type BlockTestSuite struct {
	suite.Suite
	block1000         btcjson.GetBlockVerboseTxResult
	block409008       btcjson.GetBlockVerboseTxResult
	blockHeader1000   btcjson.GetBlockHeaderVerboseResult
	blockHeader409008 btcjson.GetBlockHeaderVerboseResult
	tx409008          btcjson.TxRawResult
}

func (s *BlockTestSuite) SetupTest() {
	s.NoError(parseTestData("testdata/block_1000.json", &s.block1000))
	s.NoError(parseTestData("testdata/block_409008.json", &s.block409008))
	s.NoError(parseTestData("testdata/block_header_1000.json", &s.blockHeader1000))
	s.NoError(parseTestData("testdata/block_header_409008.json", &s.blockHeader409008))
	s.NoError(parseTestData("testdata/transaction.json", &s.tx409008))
}

func (s *BlockTestSuite) TestBlockHeaderData() {
	blk := NewBlockHeader(s.blockHeader409008)
	s.Equal(s.blockHeader409008.Hash, blk.Hash())
	s.Equal(s.blockHeader409008.Confirmations, blk.Confirmations())
	s.Equal(int64(s.blockHeader409008.Height), blk.Height())
	s.Equal(s.blockHeader409008.Version, blk.Version())
	s.Equal(s.blockHeader409008.VersionHex, blk.VersionHex())
	s.Equal(s.blockHeader409008.MerkleRoot, blk.MerkleRoot())
	s.Equal(s.blockHeader409008.Time, blk.Time())
	s.Equal(uint32(s.blockHeader409008.Nonce), blk.Nonce())
	s.Equal(s.blockHeader409008.Bits, blk.Bits())
	s.Equal(s.blockHeader409008.Difficulty, blk.Difficulty())
	s.Equal(s.blockHeader409008.PreviousHash, blk.PreviousHash())
	s.Equal(s.blockHeader409008.NextHash, blk.NextHash())
}

func (s *BlockTestSuite) TestBlockReward() {
	blk1000 := NewBlock(s.block1000)
	blk409008 := NewBlockHeader(s.blockHeader409008)
	s.EqualValues(50.0, blk1000.Reward())
	s.EqualValues(25.0, blk409008.Reward())
}

func (s *BlockTestSuite) TestBlockData() {
	blk := NewBlock(s.block409008)
	s.Equal(s.block409008.Hash, blk.Hash())
	s.Equal(s.block409008.Confirmations, blk.Confirmations())
	s.Equal(s.block409008.StrippedSize, blk.StrippedSize())
	s.Equal(s.block409008.Size, blk.Size())
	s.Equal(s.block409008.Weight, blk.Weight())
	s.Equal(s.block409008.Height, blk.Height())
	s.Equal(s.block409008.Version, blk.Version())
	s.Equal(s.block409008.VersionHex, blk.VersionHex())
	s.Equal(s.block409008.MerkleRoot, blk.MerkleRoot())
	s.Equal(s.block409008.Time, blk.Time())
	s.Equal(s.block409008.Nonce, blk.Nonce())
	s.Equal(s.block409008.Bits, blk.Bits())
	s.Equal(s.block409008.Difficulty, blk.Difficulty())
	s.Equal(s.block409008.PreviousHash, blk.PreviousHash())
	s.Equal(s.block409008.NextHash, blk.NextHash())
	s.Len(blk.Transactions(), 1962)
	for i, tx := range s.block409008.Tx {
		s.Equal(tx.Txid, blk.Transactions()[i].TxID())
		s.Equal(i, blk.Transactions()[i].Index())
	}
}

func TestBlockTestSuite(t *testing.T) {
	suite.Run(t, new(BlockTestSuite))
}

func parseTestData[T any](filename string, v *T) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}
