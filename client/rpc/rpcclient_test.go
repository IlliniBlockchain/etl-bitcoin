package rpcclient

import (
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

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
	BlockCount int64
	Client     *RPCClient
	WalletName string
	Address    btcutil.Address
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
	suite.BlockCount = 20

}

func (suite *RPCClientTestSuite) TestSanityBlockCount() {

	client := suite.Client
	blockCountReq := client.GetBlockCountAsync()
	client.Send()
	blockCount, err := blockCountReq.Receive()
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), suite.BlockCount, blockCount)
}

func TestRPCClientTestSuite(t *testing.T) {
	suite.Run(t, new(RPCClientTestSuite))
}
