package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func block_filename(height int64) string {
	return "testdata/block_" + fmt.Sprint(height) + ".json"
}

const MinBlockNumber = int64(0)
const MaxBlockNumber = int64(5)

func getTestBlockHashes(mockClient *MockClient) []*chainhash.Hash {
	hashes := make([]*chainhash.Hash, 0)
	for _, block := range mockClient.Blocks() {
		hash, err := chainhash.NewHashFromStr(block.BlockHeader.Hash())
		if err != nil {
			return nil
		}
		hashes = append(hashes, hash)
	}
	return hashes
}

type LoaderTestSuite struct {
	suite.Suite
	mockClient   *MockClient
	mockDatabase *MockDatabase
}

func (s *LoaderTestSuite) SetupTest() {
	// Parse testdata and add to test suite
	blocks := make([]*types.Block, 0)
	for i := MinBlockNumber; i <= MaxBlockNumber; i++ {
		var blockResult btcjson.GetBlockVerboseTxResult
		filename := block_filename(i)
		if err := parseTestData(filename, &blockResult); err != nil {
			s.T().Fatal(err)
		}
		var block = types.NewBlock(blockResult)
		blocks = append(blocks, block)
	}
	// Create mock client and add to test suite
	s.mockClient = NewMockClient(blocks)
	// Create mock database and add to test suite
	s.mockDatabase = NewMockDatabase()
}

func (s *LoaderTestSuite) TestBlockRangeHandler() {
	type args struct {
		client client.Client
		msg    *LoaderMsg[BlockRange]
	}

	hashes := getTestBlockHashes(s.mockClient)

	tests := []struct {
		name    string
		args    args
		want    *LoaderMsg[[]*chainhash.Hash]
		wantErr bool
	}{
		{
			name: "Test blockRangeHandler with full range",
			args: args{
				client: s.mockClient,
				msg: &LoaderMsg[BlockRange]{
					blockRange: BlockRange{
						startBlockHeight: MinBlockNumber,
						endBlockHeight:   MaxBlockNumber,
					},
					dbTx: nil,
					data: BlockRange{
						startBlockHeight: MinBlockNumber,
						endBlockHeight:   MaxBlockNumber,
					},
				},
			},
			want: &LoaderMsg[[]*chainhash.Hash]{
				blockRange: BlockRange{
					startBlockHeight: MinBlockNumber,
					endBlockHeight:   MaxBlockNumber,
				},
				dbTx: nil,
				data: hashes,
			},
			wantErr: false,
		},
		{
			name: "Test blockRangeHandler with invalid block range",
			args: args{
				client: s.mockClient,
				msg: &LoaderMsg[BlockRange]{
					blockRange: BlockRange{
						startBlockHeight: MinBlockNumber,
						endBlockHeight:   MaxBlockNumber + 1,
					},
					dbTx: nil,
					data: BlockRange{
						startBlockHeight: MinBlockNumber,
						endBlockHeight:   MaxBlockNumber + 1,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := blockRangeHandler(tt.args.client, tt.args.msg)

			if tt.wantErr {
				assert.Error(s.T(), err)
				return
			}
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.want, got)
		})
	}

}

func (s *LoaderTestSuite) TestBlockHashHandler() {
	type args struct {
		client client.Client
		msg    *LoaderMsg[[]*chainhash.Hash]
	}

	hashes := getTestBlockHashes(s.mockClient)
	blocks := s.mockClient.Blocks()

	rand.Seed(time.Now().UnixNano())
	invalidHashBytes := make([]byte, 32)
	rand.Read(invalidHashBytes)
	invalidHash, err := chainhash.NewHash(invalidHashBytes)
	s.NoError(err)

	tests := []struct {
		name    string
		args    args
		want    *LoaderMsg[[]*types.Block]
		wantErr bool
	}{
		{
			name: "Test blockHashHandler with full range",
			args: args{
				client: s.mockClient,
				msg: &LoaderMsg[[]*chainhash.Hash]{
					blockRange: BlockRange{
						startBlockHeight: MinBlockNumber,
						endBlockHeight:   MaxBlockNumber,
					},
					dbTx: nil,
					data: hashes,
				},
			},
			want: &LoaderMsg[[]*types.Block]{
				blockRange: BlockRange{
					startBlockHeight: MinBlockNumber,
					endBlockHeight:   MaxBlockNumber,
				},
				dbTx: nil,
				data: blocks,
			},
			wantErr: false,
		},
		{
			name: "Test blockHashHandler with invalid hash",
			args: args{
				client: s.mockClient,
				msg: &LoaderMsg[[]*chainhash.Hash]{
					blockRange: BlockRange{
						startBlockHeight: MinBlockNumber,
						endBlockHeight:   MaxBlockNumber,
					},
					dbTx: nil,
					data: []*chainhash.Hash{invalidHash},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := blockHashHandler(tt.args.client, tt.args.msg)

			if tt.wantErr {
				assert.Error(s.T(), err)
				return
			}
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.want, got)
		})
	}

}

func (s *LoaderTestSuite) TestBlockHandler() {
	type args struct {
		dbTx *MockDBTx
		msg  *LoaderMsg[[]*types.Block]
	}

	dbtx_full := s.mockDatabase.NewMockDBTx()
	blocks := s.mockClient.Blocks()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test blockHandler with full blocks",
			args: args{
				dbTx: dbtx_full,
				msg: &LoaderMsg[[]*types.Block]{
					blockRange: BlockRange{
						startBlockHeight: MinBlockNumber,
						endBlockHeight:   MaxBlockNumber,
					},
					dbTx: dbtx_full,
					data: blocks,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := blockHandler(tt.args.dbTx, tt.args.msg)
			if tt.wantErr {
				assert.Error(s.T(), err)
				return
			}
			assert.NoError(s.T(), err)

			correctHeaders := make([]*types.BlockHeader, len(blocks))
			correctTxs := make([]*types.Transaction, 0)
			for i, block := range tt.args.msg.data {
				correctHeaders[i] = block.BlockHeader
				correctTxs = append(correctTxs, block.Transactions()...)
			}

			assert.Equal(s.T(), correctHeaders, tt.args.dbTx.ReceivedBlockHeaders())
			assert.Equal(s.T(), correctTxs, tt.args.dbTx.ReceivedTxs())

		})
	}

}

// q: how to use contexts?
// a: https://blog.golang.org/context

func (s *LoaderTestSuite) TestLoaderManager() {
	// create a context without a cancel
	ctx := context.Background()
	// Test loader manager with full range
	loaderManager, _ := NewLoaderManager(ctx, s.mockClient, s.mockDatabase, nil)
	dbTx := s.mockDatabase.NewMockDBTx()

	blockRange := BlockRange{
		startBlockHeight: MinBlockNumber,
		endBlockHeight:   MaxBlockNumber,
	}
	loaderManager.SendInput(blockRange, dbTx)

	// wait for the loader manager to finish
	loaderManager.Close()

	// commit the dbTx
	_ = dbTx.Commit()

	// check that the dbTx has the correct data
	correctHeaders := make([]*types.BlockHeader, len(s.mockClient.Blocks()))
	correctTxs := make([]*types.Transaction, 0)
	for i, block := range s.mockClient.Blocks() {
		correctHeaders[i] = block.BlockHeader
		correctTxs = append(correctTxs, block.Transactions()...)
	}

	assert.Equal(s.T(), correctHeaders, dbTx.ReceivedBlockHeaders())
	assert.Equal(s.T(), correctTxs, dbTx.ReceivedTxs())
	assert.Equal(s.T(), dbTx.Committed(), true)
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
