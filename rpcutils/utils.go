// Copyright (c) 2017, Jonathan Chappelow
// Copyright (c) 2018, Decred Developers
// See LICENSE for details.

package rpcutils

import (
	"fmt"
	"io/ioutil"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient"
)

// RPCVersion defines the semantic versioning configuration.
type RPCVersion struct {
	Major uint32 `json:"major"`
	Minor uint32 `json:"minor"`
	Patch uint32 `json:"patch"`
}

func (v *RPCVersion) String() string {
	return fmt.Sprintf("RPC Version (V%d.%d.%d)", v.Major, v.Minor, v.Patch)
}

// ConnectRPCNode attempts to create a new websocket connection to a dcrd node,
// with the provided credentials and optional notification handlers. It also
// returns the prc server version.
func ConnectRPCNode(host, user, pass, cert string, disableTLS bool,
	ntfnHandlers ...*rpcclient.NotificationHandlers) (*rpcclient.Client,
	*RPCVersion, error) {
	var dcrdCerts []byte
	var err error
	if !disableTLS {
		dcrdCerts, err = ioutil.ReadFile(cert)
		if err != nil {
			log.Errorf("Failed to read dcrd cert file at %s: %s\n",
				cert, err.Error())
			return nil, nil, err
		}
		log.Debugf("Attempting to connect to dcrd RPC %s as user %s "+
			"using certificate located in %s",
			host, user, cert)
	} else {
		log.Debugf("Attempting to connect to dcrd RPC %s as user %s (no TLS)",
			host, user)
	}

	connCfgDaemon := &rpcclient.ConnConfig{
		Host:         host,
		Endpoint:     "ws", // websocket
		User:         user,
		Pass:         pass,
		Certificates: dcrdCerts,
		DisableTLS:   disableTLS,
	}

	var ntfnHndlrs *rpcclient.NotificationHandlers
	if len(ntfnHandlers) == 1 {
		ntfnHndlrs = ntfnHandlers[0]
	}

	if len(ntfnHandlers) != 1 || ntfnHndlrs == nil {
		log.Info("RPC notifications will not be triggered.")
	}

	dcrdClient, err := rpcclient.New(connCfgDaemon, ntfnHndlrs)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to start dcrd RPC client: %s", err.Error())
	}

	ver, err := dcrdClient.Version()
	if err != nil {
		log.Error("Unable to get RPC version: ", err)
		return nil, nil, fmt.Errorf("unable to get node RPC version")
	}
	v := ver["dcrdjsonrpcapi"]
	s := &RPCVersion{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}

	return dcrdClient, s, nil
}

// GetBlock gets a block at the given height from a chain server.
func GetBlock(client *rpcclient.Client, height int64) (*dcrutil.Block,
	*chainhash.Hash, error) {
	blockhash, err := client.GetBlockHash(height)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBlockHash(%d) failed: %v", height, err)
	}

	block, err := GetBlockByHash(client, blockhash)
	return block, blockhash, err
}

// GetBlockByHash gets the block with the given hash from a chain server.
func GetBlockByHash(client *rpcclient.Client, blockhash *chainhash.Hash) (
	*dcrutil.Block, error) {
	msgBlock, err := client.GetBlock(blockhash)
	if err != nil {
		return nil, fmt.Errorf("GetBlock failed (%s): %v", blockhash, err)
	}
	block := dcrutil.NewBlock(msgBlock)

	return block, nil
}

// GetTransactionVerboseByID get a transaction by transaction id
func GetTransactionVerboseByID(client *rpcclient.Client, txid string) (
	*dcrjson.TxRawResult, error) {
	txhash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		log.Errorf("Invalid transaction hash %s", txid)
		return nil, err
	}

	txraw, err := client.GetRawTransactionVerbose(txhash)
	if err != nil {
		log.Errorf("GetRawTransactionVerbose failed for: %v", txhash)
		return nil, err
	}
	return txraw, nil
}

// SearchRawTransaction fetch transactions that belong to the provided address
func SearchRawTransaction(client *rpcclient.Client, count int, address string) (
	[]*dcrjson.SearchRawTransactionsResult, error) {
	addr, err := dcrutil.DecodeAddress(address)
	if err != nil {
		log.Infof("Invalid address %s: %v", address, err)
		return nil, err
	}

	txs, err := client.SearchRawTransactionsVerbose(addr, 0, count,
		true, true, nil)
	if err != nil {
		log.Warnf("SearchRawTransaction failed for address %s: %v", addr, err)
	}
	return txs, nil
}
