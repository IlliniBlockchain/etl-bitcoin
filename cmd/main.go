package main

import (
	"log"

	rpc "github.com/IlliniBlockchain/etl-bitcoin/client/rpc"
	rpcclient "github.com/btcsuite/btcd/rpcclient"
)

func main() {
	client := newTestClient()
	defer client.Shutdown()
	// db := database.NewNeo4j()

	setupBlocks(*client.Client)

	// ping(client)

	// test loader stuff

	// for every block
	// get blocks and txs
	// write to disk

	// endBlockHeight := int64(25)
	// blockRangeInterval := int64(5)
	// currStart := int64(0)
	// for currStart <= endBlockHeight {
	// 	currEnd := Max(currStart+blockRangeInterval, endBlockHeight)

	// 	// getting stuff from rpc
	// 	hashes, err := loader.BlockRangesToHashes(client, int64(currStart), int64(currEnd))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	blocks, err := loader.HashesToBlocks(client, hashes)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	// transforming/extracting data
	// 	txs, err := loader.BlocksToTxs(blocks)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	for _, tx := range txs {
	// 		fmt.Println(tx.Hash())
	// 	}

	// 	currStart += blockRangeInterval
	// }
}

func setupBlocks(client rpcclient.Client) {

	WalletName := "testwallet"
	walletReq := client.CreateWalletAsync(WalletName)
	client.Send()
	_, err := walletReq.Receive()
	if err != nil {
		log.Fatal(err)
	}
	// assert.NoError(suite.T(), err)

	// Get new address
	addressReq := client.GetNewAddressAsync(WalletName)
	client.Send()
	Address, err := addressReq.Receive()
	if err != nil {
		log.Fatal(err)
	}

	// Generate blocks
	var nBlocks int64 = 30
	generateReq := client.GenerateToAddressAsync(nBlocks, Address, nil)
	client.Send()
	_, err = generateReq.Receive()
	if err != nil {
		log.Fatal(err)
	}
}

func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func newTestClient() *rpc.RPCClient {
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:18443",
		User:         "test",
		Pass:         "test",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	client, err := rpc.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func ping(client *rpc.RPCClient) {
	// Get the current block count.
	blockCountReq := client.Client.GetBlockCountAsync()
	client.Send()
	blockCount, err := blockCountReq.Receive()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)
}
