package main

import (
	"context"
	"log"

	rpc "github.com/IlliniBlockchain/etl-bitcoin/client/rpc"
	"github.com/IlliniBlockchain/etl-bitcoin/database/csv/neo4j_csv"
	"github.com/IlliniBlockchain/etl-bitcoin/loader"
	rpcclient "github.com/btcsuite/btcd/rpcclient"
)

func main() {
	client, err := rpc.New(&rpcclient.ConnConfig{
		Host:         "localhost:18443",
		User:         "test",
		Pass:         "test",
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	db, err := neo4j_csv.NewDatabase(context.Background(), map[string]interface{}{"dataDir": "./data"})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	loaderManager, err := loader.NewLoaderManager(context.Background(), client)
	if err != nil {
		log.Fatal(err)
	}
	defer loaderManager.Close()

	stats := make(chan *loader.LoaderStats)
	defer close(stats)
	go func() {
		for stat := range stats {
			stat.Wait()
			log.Print(stat)
		}
	}()

	inc := int64(10_000)
	min := int64(0)
	max := int64(300_000)
	for min < max {
		dbTx, err := db.NewDBTx()
		if err != nil {
			log.Fatal(err)
		}
		stat, err := loaderManager.SendInput(loader.BlockRange{Start: min, End: min + inc - 1}, dbTx)
		if err != nil {
			log.Fatal(err)
		}
		stats <- stat
		min += inc
	}
	if err := loaderManager.Close(); err != nil {
		log.Fatal(err)
	}
}
