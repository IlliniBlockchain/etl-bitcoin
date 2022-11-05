package rpcclient

import (
	"fmt"
	"log"

	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
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

// GetBlockHashesByRange returns block hashes from the server given a range (inclusive) of block numbers.
// Hashes are returned in order from `minBlockNumber` to `maxBlockNumber`
func (client *RPCClient) GetBlockHashesByRange(minBlockNumber, maxBlockNumber int64) (hashes []*chainhash.Hash, err error) {
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

// GetBlockHeadersByRange returns block headers from the server given a list/range of block hashes.
func (client *RPCClient) GetBlockHeaders(hashes []*chainhash.Hash) (blockHeaders []*types.BlockHeader, err error) {
	// Queue block requests
	blockReqs := make([]rpcclient.FutureGetBlockHeaderVerboseResult, len(hashes))
	for i, blockHash := range hashes {
		blockReqs[i] = client.GetBlockHeaderVerboseAsync(blockHash)
	}
	// Send
	client.Send()
	// Receive block requests
	blockHeaders = make([]*types.BlockHeader, len(hashes))
	for i, req := range blockReqs {
		block, err := req.Receive()
		if err != nil {
			return nil, err
		}
		blockHeaders[i] = types.NewBlockHeader(*block)
	}
	return blockHeaders, nil
}

// GetBlocksByRange returns blocks with transactions from the server given a list/range of block hashes.
func (client *RPCClient) GetBlocks(hashes []*chainhash.Hash) (blocks []*types.Block, err error) {
	// Queue block requests
	blockReqs := make([]rpcclient.FutureGetBlockVerboseTxResult, len(hashes))
	for i, blockHash := range hashes {
		blockReqs[i] = client.GetBlockVerboseTxAsync(blockHash)
	}
	// Send
	client.Send()
	// Receive block requests
	blocks = make([]*types.Block, len(hashes))
	for i, req := range blockReqs {
		block, err := req.Receive()
		if err != nil {
			return nil, err
		}
		blocks[i] = types.NewBlock(*block)
	}
	return blocks, nil
}
