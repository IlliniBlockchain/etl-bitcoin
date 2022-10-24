package rpcclient

import (
	"testing"

	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var _ client.Client = (*RPCClient)(nil)

func GetTestRPCClient() (*RPCClient, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:18443", // regtest
		User:         "test",
		Pass:         "test",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	client, err := New(connCfg, nil)
	return client, err
}

type RPCClientTestSuite struct {
	suite.Suite
	Client      *RPCClient
	WalletName  string
	Address     btcutil.Address
	BlockCount  int64
	BlockHashes []*chainhash.Hash
}

func (suite *RPCClientTestSuite) SetupSuite() {
	// Initialize client
	client, err := GetTestRPCClient()
	assert.NoError(suite.T(), err)
	suite.Client = client

	// Create wallet
	suite.WalletName = "testwallet"
	walletReq := client.CreateWalletAsync(suite.WalletName)
	client.Send()
	_, err = walletReq.Receive()
	assert.NoError(suite.T(), err)

	// Get new address
	addressReq := client.GetNewAddressAsync(suite.WalletName)
	client.Send()
	suite.Address, err = addressReq.Receive()
	assert.NoError(suite.T(), err)

	// Generate blocks
	var nBlocks int64 = 20
	generateReq := client.GenerateToAddressAsync(nBlocks, suite.Address, nil)
	client.Send()
	hashes, err := generateReq.Receive()
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), nBlocks, len(hashes))
	suite.BlockCount = nBlocks

	// Set block hashes
	genBlockHashReq := client.GetBlockHashAsync(0)
	client.Send()
	genBlockHash, err := genBlockHashReq.Receive()
	assert.NoError(suite.T(), err)
	suite.BlockHashes = append([]*chainhash.Hash{genBlockHash}, hashes...)
}

func (suite *RPCClientTestSuite) TestSanityBlockCount() {
	client := suite.Client
	blockCountReq := client.GetBlockCountAsync()
	client.Send()
	blockCount, err := blockCountReq.Receive()
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), suite.BlockCount, blockCount)
}

type RangeArgs struct {
	minBlockNumber int64
	maxBlockNumber int64
}

type HashTest struct {
	name    string
	args    RangeArgs
	want    []*chainhash.Hash
	wantErr bool
}

func GetHashTestTable(suite *RPCClientTestSuite) []HashTest {
	if len(suite.BlockHashes) < 5 {
		panic("Test blockchain must have 5 or more blocks.")
	}

	return []HashTest{
		{
			name: "genesis_single_block",
			args: RangeArgs{0, 0},
			want: suite.BlockHashes[0:1],
		},
		{
			name: "first_and_second_blocks",
			args: RangeArgs{1, 2},
			want: suite.BlockHashes[1:3],
		},
		{
			name: "latest_five_blocks",
			args: RangeArgs{int64(len(suite.BlockHashes) - 5), int64(len(suite.BlockHashes) - 1)},
			want: suite.BlockHashes[len(suite.BlockHashes)-5:],
		},
		{
			name:    "negative_range",
			args:    RangeArgs{-2, 2},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "reverse_range",
			args:    RangeArgs{4, 2},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "future_range",
			args:    RangeArgs{int64(len(suite.BlockHashes) - 2), int64(len(suite.BlockHashes) + 2)},
			want:    nil,
			wantErr: true,
		},
	}
}

func (suite *RPCClientTestSuite) TestGetHashesByRange() {
	tests := GetHashTestTable(suite)
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			hashes, err := suite.Client.getBlockHashesByRange(tt.args.minBlockNumber, tt.args.maxBlockNumber)
			if tt.wantErr {
				assert.Error(suite.T(), err)
				return
			}
			assert.NoError(suite.T(), err)
			assert.ElementsMatch(suite.T(), tt.want, hashes)
		})
	}
}

func (suite *RPCClientTestSuite) TestGetBlocksByRange() {
	tests := GetHashTestTable(suite)
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			blocks, err := suite.Client.GetBlocksByRange(tt.args.minBlockNumber, tt.args.maxBlockNumber)
			if tt.wantErr {
				assert.Error(suite.T(), err)
				return
			}
			assert.NoError(suite.T(), err)

			hashes := make([]*chainhash.Hash, len(blocks))
			for i, block := range blocks {
				hash := block.BlockHash()
				hashes[i] = &hash
			}
			assert.ElementsMatch(suite.T(), tt.want, hashes)
		})
	}
}

func (suite *RPCClientTestSuite) TestGetBlocksVerboseByRange() {
	tests := GetHashTestTable(suite)
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			blocks, err := suite.Client.GetBlocksVerboseByRange(tt.args.minBlockNumber, tt.args.maxBlockNumber)
			if tt.wantErr {
				assert.Error(suite.T(), err)
				return
			}
			assert.NoError(suite.T(), err)

			hashes := make([]*chainhash.Hash, len(blocks))
			for i, block := range blocks {
				hashes[i], err = chainhash.NewHashFromStr(block.Hash)
				assert.NoError(suite.T(), err)
			}
			assert.ElementsMatch(suite.T(), tt.want, hashes)
		})
	}
}

func (suite *RPCClientTestSuite) TestGetBlocksVerboseTxByRange() {
	tests := GetHashTestTable(suite)
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			blocks, err := suite.Client.GetBlocksVerboseByRange(tt.args.minBlockNumber, tt.args.maxBlockNumber)
			if tt.wantErr {
				assert.Error(suite.T(), err)
				return
			}
			assert.NoError(suite.T(), err)

			hashes := make([]*chainhash.Hash, len(blocks))
			for i, block := range blocks {
				hashes[i], err = chainhash.NewHashFromStr(block.Hash)
				assert.NoError(suite.T(), err)
			}
			assert.ElementsMatch(suite.T(), tt.want, hashes)
		})
	}
}

func TestRPCClientTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode.")
	}
	suite.Run(t, new(RPCClientTestSuite))
}
