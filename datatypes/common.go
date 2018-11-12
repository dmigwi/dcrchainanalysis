// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.
package datatypes

import (
	"time"
)

// Block describes part of the relavant block information trimmed from the block
// retrieved from the rpc client.
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

// Transaction holds generic block transaction data that is not specifically
// tied to either a transaction input or output.
type Transaction struct {
	BlockHash   string
	BlockHeight int64
	BlockTime   time.Time
	TxID        string
	TxType      int64
	TxTree      int8
	TxIndex     uint32
	Locktime    time.Time
	Expiry      time.Time
	Spent       float64
	Sent        float64
	Fees        float64
	NumInpoint  uint32
	Inpoints    []TxInput
	NumOutpoint uint32
	Outpoints   []TxOutput
}

// TxInput holds an inpoint transaction to another transaction.
type TxInput struct {
	PreviousTxHash  string
	PreviousTxIndex uint32
	PreviousTxTree  int8
	ValueIn         float64
	BlockHeight     int64
	BlockIndex      uint32
	TxHash          string
	Vout            uint32
}

// TxOutput holds an outpoint transaction to another transaction.
type TxOutput struct {
	TxHash       string
	TxIndex      uint32
	TxTree       int8
	Value        float64
	PkScript     []byte
	PkScriptData ScriptPubKeyData
}

// ScriptPubKeyData holds the public key script decoded data.
type ScriptPubKeyData struct {
	Addresses []string
	Type      string
	ReqSigs   uint32
}
