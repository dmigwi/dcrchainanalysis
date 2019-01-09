// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package rpcutils

import (
	"math"

	"github.com/decred/dcrd/blockchain/stake"
	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/txscript"
	"github.com/decred/dcrd/wire"
	"github.com/raedahgroup/dcrchainanalysis/v1/networkconfig"
)

// ExtractBlockData extracts the required block data from the MsgBlock provided.
func ExtractBlockData(blockData *wire.MsgBlock) *Block {
	blockHeader := blockData.Header
	return &Block{
		Hash:         blockData.BlockHash().String(),
		MerkleRoot:   blockHeader.MerkleRoot.String(),
		Height:       int64(blockHeader.Height),
		Time:         blockHeader.Timestamp.Unix(),
		FreshStake:   blockHeader.FreshStake,
		Voters:       blockHeader.Voters,
		StakeRoot:    blockHeader.StakeRoot.String(),
		PreviousHash: blockHeader.PrevBlock.String(),
	}
}

// ExtractBlockTransactions extracts the transactions data from the provided
// MsgBlock. Only the stake transactions data that is currently extracted.
func ExtractBlockTransactions(blockData *wire.MsgBlock,
	activeNet networkconfig.NetworkType) []*Transaction {
	block := ExtractBlockData(blockData)
	sTxs := blockData.STransactions
	txs := make([]*Transaction, len(sTxs))

	for index, sTx := range sTxs {
		tx := &Transaction{
			TxID:      sTx.TxHash().String(),
			TxType:    int64(stake.DetermineTxType(sTx)),
			TxTree:    wire.TxTreeStake,
			BlockTime: block.Time,
		}

		var sent, spent float64
		vins := make([]TxInput, len(sTx.TxIn))

		// Extract the transaction inputs.
		for v, in := range sTx.TxIn {
			vins[v] = TxInput{
				ValueIn: dcrutil.Amount(in.ValueIn).ToCoin(),
				TxHash:  tx.TxID,
			}

			spent += dcrutil.Amount(in.ValueIn).ToCoin()
		}

		tx.Spent = spent
		tx.Inpoints = vins
		tx.NumInpoint = uint32(len(vins))

		vouts := make([]TxOutput, len(sTx.TxOut))

		// Extract the transaction outputs.
		for v, out := range sTx.TxOut {
			chainParams := activeNet.ChainParams()
			scriptClass, scriptAddrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(
				out.Version, out.PkScript, chainParams)

			addys := make([]string, 0, len(scriptAddrs))
			for ia := range scriptAddrs {
				addys = append(addys, scriptAddrs[ia].String())
			}

			vouts[v] = TxOutput{
				Value:   dcrutil.Amount(out.Value).ToCoin(),
				TxIndex: uint32(v),
				PkScriptData: ScriptPubKeyData{
					Addresses: addys,
					ReqSigs:   int32(reqSigs),
					Type:      scriptClass.String(),
				},
			}

			sent += vouts[v].Value
		}

		tx.Outpoints = vouts
		tx.Sent = sent
		tx.NumOutpoint = uint32(len(vouts))
		tx.Fees = math.Round((sent-spent)*10e8) / 10e8

		txs[index] = tx
	}
	return txs
}

// ExtractRawTxTransaction extracts the transaction with all its inputs and
// outputs from a single transaction raw tx data.
func ExtractRawTxTransaction(rawTx *dcrjson.TxRawResult) *Transaction {
	tx := &Transaction{TxID: rawTx.Txid}

	var sent, spent float64
	vins := make([]TxInput, len(rawTx.Vin))

	// Extract inputs
	for v, in := range rawTx.Vin {
		vins[v] = TxInput{
			TxHash:        in.Txid,
			ValueIn:       in.AmountIn,
			OutputTxIndex: in.Vout,
		}
		sent += in.AmountIn
	}

	tx.Inpoints = vins
	tx.NumInpoint = uint32(len(vins))
	tx.Sent = sent
	tx.BlockTime = rawTx.Blocktime

	vouts := make([]TxOutput, len(rawTx.Vout))

	// Extract outputs
	for v, out := range rawTx.Vout {
		vouts[v] = TxOutput{
			Value:   out.Value,
			TxIndex: out.N,
			PkScriptData: ScriptPubKeyData{
				Addresses: out.ScriptPubKey.Addresses,
				ReqSigs:   out.ScriptPubKey.ReqSigs,
				Type:      out.ScriptPubKey.Type,
			},
		}

		spent += out.Value
	}

	tx.Outpoints = vouts
	tx.Sent = sent
	tx.NumOutpoint = uint32(len(vouts))
	tx.Fees = math.Round((sent-spent)*10e8) / 10e8
	return tx
}
