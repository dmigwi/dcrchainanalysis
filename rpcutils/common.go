// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.
package rpcutils

import (
	"time"
)

// Block describes part of the block information trimmed from the block data
// by the rpc client.
type Block struct {
	Height       int64
	Time         time.Time
	Voters       uint16
	FreshStake   uint8
	Hash         string
	MerkleRoot   string
	StakeRoot    string
	PreviousHash string
}

// Transaction holds generic block transaction data that contains both the
// transaction's input or output data.
type Transaction struct {
	TxID        string
	TxType      int64
	TxTree      int8
	Spent       float64
	Sent        float64
	Fees        float64
	NumInpoint  uint32
	Inpoints    []TxInput
	NumOutpoint uint32
	Outpoints   []TxOutput
}

// TxInput holds an inpoint transaction of a given transaction.
type TxInput struct {
	ValueIn       float64
	TxHash        string
	OutputTxIndex uint32
}

// TxOutput holds an outpoint transaction of a given transaction.
type TxOutput struct {
	Value        float64
	TxIndex      uint32
	PkScriptData ScriptPubKeyData
}

// ScriptPubKeyData holds the public key script decoded data.
type ScriptPubKeyData struct {
	Addresses []string
	Type      string
	ReqSigs   int32
}
