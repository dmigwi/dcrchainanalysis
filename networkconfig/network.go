// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package networkconfig

import (
	"github.com/decred/dcrd/chaincfg"
)

const (
	// MainNet is the environment where the decred actual blockchain runs on.
	MainNet NetworkType = iota

	// TestNet is a smaller version of decred net/environment that is used to
	// help run tests of features that alter the contents of the blockchain.
	TestNet

	// SimNet is an environment smaller than TestNet that is typically a few blocks
	// long and is used to test and simulate decred applications than require the
	// minimum number of blocks in a blockchain to run.
	SimNet
)

// NetworkType defines the various Decred networks currently supported.
type NetworkType int

type networkParams struct {
	Name    string
	RPCPort string
	Params  *chaincfg.Params
}

var (
	mainnetParams = networkParams{
		RPCPort: "9109",
		Name:    "Mainnet",
		Params:  &chaincfg.MainNetParams,
	}

	testnetParams = networkParams{
		RPCPort: "19108",
		Name:    "Testnet",
		Params:  &chaincfg.TestNet3Params,
	}

	simnetParams = networkParams{
		RPCPort: "19556",
		Name:    "Simnet",
		Params:  &chaincfg.SimNetParams,
	}
)

// RPCPort returns the default RPC port value supported for the given net.
func (n NetworkType) RPCPort() string {
	switch n {
	case TestNet:
		return testnetParams.RPCPort
	case SimNet:
		return simnetParams.RPCPort
	default:
		return mainnetParams.RPCPort
	}
}

func (n NetworkType) String() string {
	switch n {
	case TestNet:
		return testnetParams.Name
	case SimNet:
		return simnetParams.Name
	default:
		return mainnetParams.Name
	}
}

// ChainParams returns the chain parameters associated with the selected network.
func (n NetworkType) ChainParams() *chaincfg.Params {
	switch n {
	case TestNet:
		return testnetParams.Params
	case SimNet:
		return simnetParams.Params
	default:
		return mainnetParams.Params
	}
}
