// +build testnet

package rpcutils

import (
	"testing"
)

const (
	// Test RPC connection configuration
	rpcUser     = "dcrd"
	rpcPassword = "dcrd"
	host        = "127.0.0.1:19109"
)

// TestConnectRPCNode tests the functionality of ConnectRPCNode function.
func TestConnectRPCNode(t *testing.T) {
	client, v, err := ConnectRPCNode(host, rpcUser, rpcPassword, "", true)
	t.Run("TestConnectRPCNode", func(t *testing.T) {
		if err != nil {
			t.Fatalf("Expected error to be nil but found %v", err)
		}

		if v == nil {
			t.Fatal("Expected the RPC JSON version not to be nil but it was nil.")
		}

		if client == nil {
			t.Fatal("Expected the client not to be nil but it was nil.")
		}
	})
}
