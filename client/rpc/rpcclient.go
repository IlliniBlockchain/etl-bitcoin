package rpcclient

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

// RPCClient represents a JSON RPC connection to a bitcoin node.
type RPCClient struct {
	*rpcclient.Client
}

// New acts as a default constructor for our RPCClient extending functionality of btcd/rpcclient.Client
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

// getBlockHashesByRange returns block hashes from the server given a range (inclusive) of block numbers.
// Hashes are returned in order from `minBlockNumber` to `maxBlockNumber`
func (client *RPCClient) getBlockHashesByRange(minBlockNumber, maxBlockNumber int64) (hashes []*chainhash.Hash, err error) {
	if minBlockNumber > maxBlockNumber {
		log.Printf("minBlockNumber: %d\tmaxBlockNumber: %d\n", minBlockNumber, maxBlockNumber)
		return nil, fmt.Errorf(
			"minBlockNumber (%d) must be less than or equal to maxBlockNumber (%d)",
			minBlockNumber,
			maxBlockNumber,
		)
	}
	nBlocks := maxBlockNumber - minBlockNumber + 1

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

	return blockHashes, nil
}

// getBlocksByHashes returns raw blocks from the server given a list of block hashes.
func (client *RPCClient) getBlocksByHashes(hashes []*chainhash.Hash) (blocks []*wire.MsgBlock, err error) {
	// Queue block requests
	blockReqs := make([]rpcclient.FutureGetBlockResult, len(hashes))
	for i, blockHash := range hashes {
		blockReqs[i] = client.GetBlockAsync(blockHash)
	}
	// Send
	client.Send()
	// Receive block requests
	blocks = make([]*wire.MsgBlock, len(hashes))
	for i, req := range blockReqs {
		blocks[i], err = req.Receive()
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

// getBlocksVerboseByHashes returns data structures from the server with information
// about block given a list of block hashes.
func (client *RPCClient) getBlocksVerboseByHashes(hashes []*chainhash.Hash) (blocks []*btcjson.GetBlockVerboseResult, err error) {
	// Queue block requests
	blockReqs := make([]rpcclient.FutureGetBlockVerboseResult, len(hashes))
	for i, blockHash := range hashes {
		blockReqs[i] = client.GetBlockVerboseAsync(blockHash)
	}
	// Send
	client.Send()
	// Receive block requests
	blocks = make([]*btcjson.GetBlockVerboseResult, len(hashes))
	for i, req := range blockReqs {
		blocks[i], err = req.Receive()
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

// getBlocksVerboseTxByHashes returns data structures from the server with information
// about blocks and their transactions given a list of block hashes.
func (client *RPCClient) getBlocksVerboseTxByHashes(hashes []*chainhash.Hash) (blocks []*btcjson.GetBlockVerboseTxResult, err error) {
	// Queue block requests
	blockReqs := make([]rpcclient.FutureGetBlockVerboseTxResult, len(hashes))
	for i, blockHash := range hashes {
		blockReqs[i] = client.GetBlockVerboseTxAsync(blockHash)
	}
	// Send
	client.Send()
	// Receive block requests
	blocks = make([]*btcjson.GetBlockVerboseTxResult, len(hashes))
	for i, req := range blockReqs {
		blocks[i], err = req.Receive()
		if err != nil {
			return nil, err
		}
	}
	return blocks, nil
}

// GetBlocksByRange returns raw blocks from the server given a range (inclusive) of block numbers.
func (client *RPCClient) GetBlocksByRange(minBlockNumber, maxBlockNumber int64) (blocks []*wire.MsgBlock, err error) {
	blockHashes, err := client.getBlockHashesByRange(minBlockNumber, maxBlockNumber)
	if err != nil {
		return nil, err
	}
	blocks, err = client.getBlocksByHashes(blockHashes)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

// GetBlocksVerboseByRange returns data structures from the server with information
// about block given a range of block numbers.
func (client *RPCClient) GetBlocksVerboseByRange(minBlockNumber, maxBlockNumber int64) (blocks []*btcjson.GetBlockVerboseResult, err error) {
	hashes, err := client.getBlockHashesByRange(minBlockNumber, maxBlockNumber)
	if err != nil {
		return nil, err
	}
	blocks, err = client.getBlocksVerboseByHashes(hashes)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

// GetBlocksVerboseTxByRange returns data structures from the server with information
// about blocks and their transactions given a range of block numbers.
func (client *RPCClient) GetBlocksVerboseTxByRange(minBlockNumber, maxBlockNumber int64) (blocks []*btcjson.GetBlockVerboseTxResult, err error) {
	hashes, err := client.getBlockHashesByRange(minBlockNumber, maxBlockNumber)
	if err != nil {
		return nil, err
	}
	blocks, err = client.getBlocksVerboseTxByHashes(hashes)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}
