package rpcclient

import "github.com/btcsuite/btcd/rpcclient"

// RPCClient represents a JSON RPC connection to a bitcoin node.
type RPCClient struct {
	*rpcclient.Client
}
