package rpcclient

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/IlliniBlockchain/etl-bitcoin/client"
	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/stretchr/testify/suite"
)

var _ client.Client = (*RPCClientPool)(nil)

func config() *rpcclient.ConnConfig {
	return &rpcclient.ConnConfig{
		Host:         "localhost:18443", // regtest
		User:         "test",
		Pass:         "test",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
}

type RPCClientTestSuite struct {
	suite.Suite
	Pool        *RPCClientPool
	Client      *rpcclient.Client
	WalletName  string
	Address     btcutil.Address
	BlockCount  int64
	BlockHashes []*chainhash.Hash
}

func (suite *RPCClientTestSuite) SetupSuite() {
	// Initialize client
	client, err := rpcclient.NewBatch(config())
	suite.NoError(err)
	suite.Client = client
	suite.Pool, err = New(config())
	suite.NoError(err)

	// Create wallet
	suite.WalletName = "testwallet"
	walletReq := client.CreateWalletAsync(suite.WalletName)
	client.Send()
	_, err = walletReq.Receive()
	suite.NoError(err)

	// Get new address
	addressReq := client.GetNewAddressAsync(suite.WalletName)
	client.Send()
	suite.Address, err = addressReq.Receive()
	suite.NoError(err)

	// Generate blocks
	var nBlocks int64 = 20
	generateReq := client.GenerateToAddressAsync(nBlocks, suite.Address, nil)
	client.Send()
	hashes, err := generateReq.Receive()
	suite.NoError(err)
	suite.EqualValues(nBlocks, len(hashes))
	suite.BlockCount = nBlocks

	// Set block hashes
	genBlockHashReq := client.GetBlockHashAsync(0)
	client.Send()
	genBlockHash, err := genBlockHashReq.Receive()
	suite.NoError(err)
	suite.BlockHashes = append([]*chainhash.Hash{genBlockHash}, hashes...)
}

func (suite *RPCClientTestSuite) TestSanityBlockCount() {
	client := suite.Client
	blockCountReq := client.GetBlockCountAsync()
	client.Send()
	blockCount, err := blockCountReq.Receive()
	suite.NoError(err)
	suite.EqualValues(suite.BlockCount, blockCount)
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

type BlockHashTest struct {
	name    string
	args    []*chainhash.Hash
	want    []*chainhash.Hash
	wantErr bool
}

func GetBlockHashTestTable(suite *RPCClientTestSuite) []BlockHashTest {
	rand.Seed(time.Now().UnixNano())
	invalidHashBytes := make([]byte, 32)
	rand.Read(invalidHashBytes)
	invalidHash, _ := chainhash.NewHash(invalidHashBytes)

	return []BlockHashTest{
		{
			name: "genesis_single_block",
			args: suite.BlockHashes[0:1],
			want: suite.BlockHashes[0:1],
		},
		{
			name: "first_and_second_blocks",
			args: suite.BlockHashes[1:3],
			want: suite.BlockHashes[1:3],
		},
		{
			name: "latest_five_blocks",
			args: suite.BlockHashes[len(suite.BlockHashes)-5:],
			want: suite.BlockHashes[len(suite.BlockHashes)-5:],
		},
		{
			name: "nonsequential_blocks",
			args: []*chainhash.Hash{suite.BlockHashes[2], suite.BlockHashes[0], suite.BlockHashes[5], suite.BlockHashes[4]},
			want: []*chainhash.Hash{suite.BlockHashes[2], suite.BlockHashes[0], suite.BlockHashes[5], suite.BlockHashes[4]},
		},
		{
			name:    "invalid_hash",
			args:    []*chainhash.Hash{suite.BlockHashes[3], invalidHash},
			want:    nil,
			wantErr: true,
		},
	}
}

func (suite *RPCClientTestSuite) TestGetHashesByRange() {
	tests := GetHashTestTable(suite)
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			hashes, err := suite.Pool.GetBlockHashesByRange(tt.args.minBlockNumber, tt.args.maxBlockNumber)
			if tt.wantErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)
			suite.ElementsMatch(tt.want, hashes)
		})
	}
}

func (suite *RPCClientTestSuite) TestGetBlockHeaders() {
	tests := GetBlockHashTestTable(suite)
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			blockHeaders, err := suite.Pool.GetBlockHeaders(tt.args)
			if tt.wantErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)

			blockHeaderHashes := make([]*chainhash.Hash, len(blockHeaders))
			for i, block := range blockHeaders {
				hash, err := chainhash.NewHashFromStr(block.Hash())
				suite.NoError(err)
				blockHeaderHashes[i] = hash
			}
			suite.ElementsMatch(tt.want, blockHeaderHashes)
		})
	}
}

func (suite *RPCClientTestSuite) TestGetBlocks() {
	tests := GetBlockHashTestTable(suite)
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			blocks, err := suite.Pool.GetBlocks(tt.args)
			if tt.wantErr {
				suite.Error(err)
				return
			}
			suite.NoError(err)

			blockHashes := make([]*chainhash.Hash, len(blocks))
			for i, block := range blocks {
				hash, err := chainhash.NewHashFromStr(block.Hash())
				suite.NoError(err)
				blockHashes[i] = hash
			}
			suite.ElementsMatch(tt.want, blockHashes)
		})
	}
}

func (suite *RPCClientTestSuite) TestPipeline() {
	hashesCh := make(chan []*chainhash.Hash, 1)
	blocksCh := make(chan *types.Block, 1)
	go func() {
		defer close(hashesCh)
		for i := 0; i < int(suite.BlockCount)-1; i += 2 {
			hashes, err := suite.Pool.GetBlockHashesByRange(int64(i), int64(i+1))
			if err != nil {
				suite.T().Error(err)
				return
			}
			hashesCh <- hashes
		}
	}()
	go func() {
		defer close(blocksCh)
		for hashes := range hashesCh {
			blocks, err := suite.Pool.GetBlocks(hashes)
			if err != nil {
				suite.T().Error(err)
				return
			}
			for _, block := range blocks {
				blocksCh <- block
			}
		}
	}()
	curBlock := int64(0)
	for block := range blocksCh {
		suite.T().Logf("received block: %d", block.Height())
		suite.Equal(curBlock, block.Height())
		suite.Equal(suite.BlockHashes[curBlock].String(), block.Hash())
		curBlock++
	}
}

func TestRPCClientTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping RPCClientTestSuite in short mode.")
	}
	suite.Run(t, new(RPCClientTestSuite))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func BenchmarkRPCClientGetBlocks(b *testing.B) {
	client, err := rpcclient.New(config(), nil)
	if err != nil {
		b.Fatal(err)
	}
	pool, err := New(config())
	if err != nil {
		b.Fatal(err)
	}

	walletName := "testwallet"
	client.CreateWallet(walletName)
	// if _, err = walletReq.Receive(); err != nil {
	// 	b.Fatal(err)
	// }
	address, err := client.GetNewAddress(walletName)
	if err != nil {
		b.Fatal(err)
	}
	hashes, err := client.GenerateToAddress(10_000, address, nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for _, total := range []int{1, 10, 100, 1_000, 10_000} {
		for _, batch := range []int{1, 10, 100, 1_000, 10_000} {
			if batch > total {
				continue
			}
			b.Run(fmt.Sprintf("total=%d,batch=%d", total, batch), func(b *testing.B) {
				for i := 0; i < b.N; i += batch {
					if _, err := pool.GetBlocks(hashes[i:min(i+batch, b.N)]); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
