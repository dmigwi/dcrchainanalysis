// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package analytics

import (
	"fmt"

	"github.com/decred/dcrd/rpcclient"
	"github.com/raedahgroup/dcrchainanalysis/v1/rpcutils"
)

// RetrieveTxData fetches a transaction data from the rpc client returns
// processed transaction data.
func RetrieveTxData(client *rpcclient.Client, txHash string) (*rpcutils.Transaction, error) {
	// Return an empty Transactions object if txHash used is empty.
	if txHash == "" {
		return &rpcutils.Transaction{}, nil
	}

	log.Infof("Retrieving data for transaction: %s", txHash)

	txData, err := rpcutils.GetTransactionVerboseByID(client, txHash)
	if err != nil {
		return nil, fmt.Errorf("RetrieveTxData error: failed to fetch transaction %s", txHash)
	}

	return rpcutils.ExtractRawTxTransaction(txData), nil
}

// RetrieveTxProbability returns the tx level probability values for each output.
func RetrieveTxProbability(client *rpcclient.Client, txHash string) (
	[]*FlowProbability, *rpcutils.Transaction, error) {
	tx, err := RetrieveTxData(client, txHash)
	if err != nil {
		return nil, nil, err
	}

	rawSolution, inputs, outputs, err := TransactionFundsFlow(tx)
	if err != nil {
		return nil, nil, err
	}

	return TxFundsFlowProbability(rawSolution, inputs, outputs), tx, nil
}

// ChainDiscovery returns all the possible chains associated with the tx hash used.
func ChainDiscovery(client *rpcclient.Client, txHash string, outputIndex ...int) ([]*Hub, error) {
	tx, err := RetrieveTxData(client, txHash)
	if err != nil {
		return nil, err
	}

	// hubsChain defines the various paths with funds flows from a given output to
	// back in time when the source for each path can be identified.
	var hubsChain []*Hub

	var outPoints []rpcutils.TxOutput

	var depth = 3

	switch {
	// OutputIndex has been provided
	case len(outputIndex) > 0:
		var txIndex int

		if outputIndex[0] > len(tx.Outpoints) {
			txIndex = len(tx.Outpoints) - 1

		} else if outputIndex[0] > 0 {
			txIndex = outputIndex[0]
		}

		outPoints = append(outPoints, tx.Outpoints[txIndex])

	// OutputIndex has not been provided.
	case len(outputIndex) == 0:
		outPoints = tx.Outpoints
	}

	for _, val := range outPoints {
		var hubCount int
		var stackTrace []*Hub
		count := 1

		entry := &Hub{
			TxHash:  tx.TxID,
			Amount:  val.Value,
			Address: val.PkScriptData.Addresses[0],
		}

		err = handleDepths(entry, stackTrace, client, count, depth, hubCount)
		if err != nil {
			return nil, err
		}

		hubsChain = append(hubsChain, entry)
	}

	log.Info("Finished auto chain(s) discovery and appending all needed data")

	return hubsChain, nil
}

// handleDepths recusively creates a graph-like data structure that shows the
// funds flow path from output (UTXO) to the source of funds at the provided depth.
func handleDepths(curHub *Hub, stack []*Hub, client *rpcclient.Client,
	count, depth, hubCount int) error {
	err := curHub.getDepth(client)
	if err != nil {
		return err
	}

	if depth == count {
		// backtrack till we find an unprocessed Hub.
		for {
			count--
			curHub = stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if hubCount+1 < len(curHub.Matched[curHub.setCount].Inputs) {
				hubCount++
				break

			} else if curHub.setCount+1 < len(curHub.Matched) {
				curHub.setCount++
				hubCount = 0
				break
			}

			if len(stack) == 0 {
				return nil
			}
		}
	}

	// Adds items to the stack.
	stack = append(stack, curHub)
	curHub = curHub.Matched[curHub.setCount].Inputs[hubCount]

	return handleDepths(curHub, stack, client, count+1, depth, hubCount)
}

// getDepth appends all the sets linked to a given output after a given amount
// probability solution is resolved.
func (h *Hub) getDepth(client *rpcclient.Client) error {
	if h.TxHash == "" {
		return nil
	}

	probability, tx, err := RetrieveTxProbability(client, h.TxHash)
	if err != nil {
		return err
	}

	for _, item := range probability {
		if item.OutputAmount == h.Amount {
			for _, entry := range item.ProbableInputs {
				d, err := getSet(client, tx, entry)
				if err != nil {
					return err
				}

				h.Probability = item.LinkingProbability
				h.Matched = append(h.Matched, d)
			}
		}
	}
	return nil
}

// The Set returned in a given output probability solution does not have a lot of
// data, this functions reconstructs the Set adding the necessary information.
func getSet(client *rpcclient.Client, txData *rpcutils.Transaction,
	matchedInputs *InputSets) (set Set, err error) {
	inputs := make([]rpcutils.TxInput, len(txData.Inpoints))
	copy(inputs, txData.Inpoints)

	for _, item := range matchedInputs.Set {
		for i := 0; i < item.PossibleInputs; i++ {
			for k, d := range inputs {
				if d.ValueIn == item.Amount {
					s := &Hub{Amount: d.ValueIn, TxHash: d.TxHash}

					tx, err := RetrieveTxData(client, s.TxHash)
					if err != nil {
						return Set{}, err
					}

					// fetch the current hub's Address.
					for k := range tx.Outpoints {
						if d.OutputTxIndex == tx.Outpoints[k].TxIndex {
							s.Address = tx.Outpoints[k].PkScriptData.Addresses[0]
						}
					}

					set.Inputs = append(set.Inputs, s)
					set.PercentOfInputs = matchedInputs.PercentOfInputs

					copy(inputs[k:], inputs[k+1:])
					inputs = inputs[:len(inputs)-1]
					break
				}
			}
		}
	}
	return
}
