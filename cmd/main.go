package main

import (
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hello world!")
	// Load username and password from .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error occurred loading .env: Err: %s", err)
	}
	user := os.Getenv("RPCUSER")
	pass := os.Getenv("RPCPASS")
	// fmt.Print(user, pass)
	// Connect to local bitcoin core RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         "localhost:8332",
		User:         user,
		Pass:         pass,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	// Get the current block count.
	blockCount, err := client.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)
}
