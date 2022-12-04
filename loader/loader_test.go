package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/suite"
)

type LoaderTestSuite struct {
	suite.Suite
	client           client.Client
	startBlockHeight int64
	endBlockHeight   int64
	blocks           []*types.Block
}

// mock client
type MockClient struct {
	blocks []*types.Block
}

func NewMockClient(blocks []*types.Block) *MockClient {
	return &MockClient{
		blocks: blocks,
	}
}

func (c *MockClient) GetBlockHashesByRange(minBlockNumber, maxBlockNumber int64) ([]*chainhash.Hash, error) {
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
		for _, block := range c.blocks {
			if block.BlockHeader.Hash() == hash.String() {
				headers = append(headers, block.BlockHeader)
			}
		}
	}
	return headers, nil
}

func (c *MockClient) GetBlocks(hashes []*chainhash.Hash) ([]*types.Block, error) {
	// get blocks by searching for the block
	blocks := make([]*types.Block, 0)
	for _, hash := range hashes {
		for _, block := range c.blocks {
			if block.BlockHeader.Hash() == hash.String() {
				blocks = append(blocks, block)
			}
		}
	}
	return blocks, nil
}

func block_filename(height int64) string {
	return "testdata/block_" + fmt.Sprint(height) + ".json"
}

func (s *LoaderTestSuite) SetupTest() {
	// Parse testdata and add to test suite
	s.startBlockHeight = 0
	s.endBlockHeight = 5
	s.blocks = make([]*types.Block, 0)
	for i := s.startBlockHeight; i <= s.endBlockHeight; i++ {
		var block types.Block
		filename := block_filename(i)
		if err := parseTestData(filename, &block); err != nil {
			s.T().Fatal(err)
		}
		s.blocks = append(s.blocks, &block)
	}
	// Create mock client and add to test suite
	s.client = NewMockClient(s.blocks)
}

func (s *LoaderTestSuite) TestBlockRangeHandler(t *testing.T) {
	type args struct {
		client client.Client
		msg    *LoaderMsg[BlockRange]
	}

	// type args struct {
	// 	client client.Client
	// 	msg    *LoaderMsg[BlockRange]
	// }
	// tests := []struct {
	// 	name    string
	// 	args    args
	// 	want    *LoaderMsg[[]*chainhash.Hash]
	// 	wantErr bool
	// }{
	// 	{
	// 		name: "return option",
	// 		args: args{
	// 			client: nil,
	// 			msg: nil,
	// 		},
	// 		want: 1,
	// 	},
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, err :=
	// 		if tt.wantErr {
	// 			assert.Error(t, err)
	// 			return
	// 		}
	// 		assert.Equal(t, tt.want, got)
	// 	})
	// }

}

func (s *LoaderTestSuite) TestBlockHashHandler(t *testing.T) {

}

func (s *LoaderTestSuite) TestBlockHandler(t *testing.T) {

}

// Requires integration test or dummy client
func (s *LoaderTestSuite) TestLoaderManager(t *testing.T) {

}

func TestLoaderTestSuite(t *testing.T) {
	suite.Run(t, new(LoaderTestSuite))
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
