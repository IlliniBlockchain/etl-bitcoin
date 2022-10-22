package rpcclient

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

// RPCClient represents a JSON RPC connection to a bitcoin node.
type RPCClient struct {
	*rpcclient.Client
}

func New(config *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) (*RPCClient, error) {
	internal_client, err := rpcclient.NewBatch(config)
	if err != nil {
		return nil, err
	}
	client := RPCClient{
		internal_client,
	}
	return &client, nil
}

// GetBlocksByRange returns raw blocks from the server given a range (inclusive) of block numbers.
func (client *RPCClient) GetBlocksByRange(minBlockNumber, maxBlockNumber int64) (blocks []*wire.MsgBlock, err error) {
	nBlocks := maxBlockNumber - minBlockNumber + 1

	// Block hashes
	// Queue block hash requests
	hashReqs := make([]rpcclient.FutureGetBlockHashResult, nBlocks)
	for i := range hashReqs {
		hashReqs[i] = client.GetBlockHashAsync(minBlockNumber + int64(i))
	}
	// Send
	client.Send()
	// Receive block hash requests
	blockHashes := make([]*chainhash.Hash, nBlocks)
	for i, req := range hashReqs {
		blockHashes[i], err = req.Receive()
		if err != nil {
			return nil, err
		}
	}

	// Blocks
	// Queue block requests
	blockReqs := make([]rpcclient.FutureGetBlockResult, nBlocks)
	for i, blockHash := range blockHashes {
		blockReqs[i] = client.GetBlockAsync(blockHash)
	}
	// Send
	client.Send()
	// Receive block requests
	blocks = make([]*wire.MsgBlock, nBlocks)
	for i, req := range blockReqs {
		blocks[i], err = req.Receive()
		if err != nil {
			return nil, err
		}
	}

	return blocks, nil
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
