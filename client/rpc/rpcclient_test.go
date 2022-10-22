package rpcclient

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/stretchr/testify/assert"
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

func TestNewRPCClient(t *testing.T) {
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:8332",
		User:         "user",
		Pass:         "pass",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	_, err := New(connCfg, nil)
	assert.NoError(t, err)
}

func TestRPCClientPing(t *testing.T) {
	client, err := GetTestRPCClient()
	assert.NoError(t, err)
	block_count, err := client.GetBlockCount()
	assert.NoError(t, err)
	fmt.Println(block_count)
}
