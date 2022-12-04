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

	db, err := neo4j_csv.NewDatabase(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	loaderManager, err := loader.NewLoaderManager(context.Background(), client, db, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer loaderManager.Close()

	loaderManager.SendInput(loader.BlockRange{Start: 0, End: 100}, nil)
}
