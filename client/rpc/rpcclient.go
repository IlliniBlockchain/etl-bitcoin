package rpcclient

import (
	"github.com/btcsuite/btcd/rpcclient"
)

// RPCClient represents a JSON RPC connection to a bitcoin node.
type RPCClient struct {
	*rpcclient.Client
}

func New(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (*RPCClient, error) {
	internal_client, err := rpcclient.New(config, nil)
	if err != nil {
		return nil, err
	}
	client := RPCClient{
		internal_client,
	}
	tries := 1
	err = client.Connect(tries)
	if err != nil {
		return nil, err
	}
	return &client, nil
}
