package rpcclient

import (
	"fmt"
	"sync"

	"github.com/IlliniBlockchain/etl-bitcoin/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"golang.org/x/sync/semaphore"
)

type clientConnection struct {
	*rpcclient.Client
	sem *semaphore.Weighted
}

func newClientConnection(config *rpcclient.ConnConfig) (*clientConnection, error) {
	client, err := rpcclient.NewBatch(config)
	if err != nil {
		return nil, err
	}
	return &clientConnection{client, semaphore.NewWeighted(1)}, nil
}

func (c *clientConnection) acquire() error {
	if !c.sem.TryAcquire(1) {
		return fmt.Errorf("clientConnection is busy")
	}
	return nil
}

func (c *clientConnection) release() {
	c.sem.Release(1)
}

func (c *clientConnection) close() error {
	c.Client.Shutdown()
	c.Client.WaitForShutdown()
	return nil
}

// RPCClientPool represents a JSON RPC connection to a bitcoin node.
type RPCClientPool struct {
	config     *rpcclient.ConnConfig
	clientPool []*clientConnection
	poolMu     sync.Mutex
}

// New acts as a default constructor for our RPCClientPool extending functionality of btcd/rpcclient.Client
func New(config *rpcclient.ConnConfig) (*RPCClientPool, error) {
	client := RPCClientPool{
		config,
		make([]*clientConnection, 0),
		sync.Mutex{},
	}
	return &client, nil
}

func (pool *RPCClientPool) acquireClient() (*clientConnection, error) {
	pool.poolMu.Lock()
	defer pool.poolMu.Unlock()
	for _, client := range pool.clientPool {
		if err := client.acquire(); err == nil {
			return client, nil
		}
	}
	client, err := newClientConnection(pool.config)
	if err != nil {
		return nil, err
	}
	if err := client.acquire(); err != nil {
		return nil, err
	}
	pool.clientPool = append(pool.clientPool, client)
	return client, nil
}

func (pool *RPCClientPool) Close() error {
	pool.poolMu.Lock()
	defer pool.poolMu.Unlock()
	for _, client := range pool.clientPool {
		client.close()
	}
	pool.clientPool = make([]*clientConnection, 0)
	return nil
}

// GetBlockHashesByRange returns block hashes from the server given a range (inclusive) of block numbers.
// Hashes are returned in order from `minBlockNumber` to `maxBlockNumber`
func (pool *RPCClientPool) GetBlockHashesByRange(minBlockNumber, maxBlockNumber int64) (hashes []*chainhash.Hash, err error) {
	if minBlockNumber > maxBlockNumber {
		return nil, fmt.Errorf(
			"minBlockNumber (%d) must be less than or equal to maxBlockNumber (%d)",
			minBlockNumber,
			maxBlockNumber,
		)
	}
	nBlocks := maxBlockNumber - minBlockNumber + 1

	// Queue block hash requests
	hashReqs := make([]rpcclient.FutureGetBlockHashResult, nBlocks)
	client, err := pool.acquireClient()
	if err != nil {
		return nil, err
	}
	for i := range hashReqs {
		hashReqs[i] = client.GetBlockHashAsync(minBlockNumber + int64(i))
	}
	// Send
	err = client.Send()
	client.release()
	if err != nil {
		return nil, err
	}
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
func (pool *RPCClientPool) GetBlockHeaders(hashes []*chainhash.Hash) (blockHeaders []*types.BlockHeader, err error) {
	// Queue block requests
	blockReqs := make([]rpcclient.FutureGetBlockHeaderVerboseResult, len(hashes))
	client, err := pool.acquireClient()
	if err != nil {
		return nil, err
	}
	for i, blockHash := range hashes {
		blockReqs[i] = client.GetBlockHeaderVerboseAsync(blockHash)
	}
	// Send
	err = client.Send()
	client.release()
	if err != nil {
		return nil, err
	}
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
func (pool *RPCClientPool) GetBlocks(hashes []*chainhash.Hash) (blocks []*types.Block, err error) {
	// Queue block requests
	blockReqs := make([]rpcclient.FutureGetBlockVerboseTxResult, len(hashes))
	client, err := pool.acquireClient()
	if err != nil {
		return nil, err
	}
	for i, blockHash := range hashes {
		blockReqs[i] = client.GetBlockVerboseTxAsync(blockHash)
	}
	// Send
	err = client.Send()
	client.release()
	if err != nil {
		return nil, err
	}
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
