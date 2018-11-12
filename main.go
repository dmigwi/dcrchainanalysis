// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/raedahgroup/dcrchainanalysis/v1/datatypes"
	"github.com/raedahgroup/dcrchainanalysis/v1/rpcutils"
)

const (
	blockHeight  int64 = 50024
	transactionX       = "d2656cf5fe1279a5a51d82820db47faff470a6bcec80692fd3629427e17699a3"
)

func start() error {
	cfg, otherConfig, err := loadConfig()
	if err != nil {
		return err
	}

	client, rpcVersion, err := rpcutils.ConnectRPCNode(cfg.DcrdServ, cfg.DcrdUser,
		cfg.DcrdPass, cfg.DcrdCert, cfg.DisableDaemonTLS, nil)
	if err != nil {
		return err
	}

	log.Infof("Connected to a dcrd node successfully: %s", rpcVersion.String())

	log.Infof("Fetching block at height %d", blockHeight)

	blockData, _, err := rpcutils.GetBlock(client, blockHeight)
	if err != nil {
		return fmt.Errorf("failed to fetch block at height %d", blockHeight)
	}

	txs := datatypes.ExtractBlockTransactions(blockData.MsgBlock(), otherConfig.ActiveNet)
	d, _ := json.Marshal(txs)

	log.Infof("All the stake Transactions associated with block %d include %s",
		blockHeight, string(d))

	log.Infof("\n\n Fetching transaction %s", transactionX)

	txData, err := rpcutils.GetTransactionVerboseByID(client, transactionX)
	if err != nil {
		return fmt.Errorf("failed to fetch transaction %s", transactionX)
	}

	tx := datatypes.ExtractRawTxTransaction(txData)
	s, _ := json.Marshal(tx)

	log.Infof("Transaction fetched has the following data %s", string(s))

	return nil
}

func main() {
	err := start()
	if err != nil {
		log.Error(err)
	}
}
