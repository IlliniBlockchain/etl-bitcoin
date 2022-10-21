package rpcclient

import (
	"testing"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/stretchr/testify/assert"
)

func TestCreateRPCClientPing(t *testing.T) {
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
