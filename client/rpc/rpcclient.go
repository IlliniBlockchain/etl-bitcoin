package rpcclient

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
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
	return &client, nil
}

func (client *RPCClient) GetBlocksByRange(minBlockNumber, maxBlockNumber int64) ([]*wire.MsgBlock, error) {
	return nil, nil
}

// GetBlocksVerboseByRange returns data structures from the server with information
// about block given a range of block numbers.
func (client *RPCClient) GetBlocksVerboseByRange(minBlockNumber, maxBlockNumber int64) ([]*btcjson.GetBlockVerboseResult, error) {
	return nil, nil
}

// GetBlocksVerboseTxByRange returns data structures from the server with information
// about blocks and their transactions given a range of block numbers.
func (client *RPCClient) GetBlocksVerboseTxByRange(minBlockNumber, maxBlockNumber int64) ([]*btcjson.GetBlockVerboseTxResult, error) {
	return nil, nil
}
