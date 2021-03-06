// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package analytics

import (
	"errors"
	"fmt"
	"sort"

	"github.com/decred/slog"
	"github.com/raedahgroup/dcrchainanalysis/v1/rpcutils"
)

// TransactionFundsFlow calculates the funds flow between a set of inputs and
// their corresponding set of outputs for the provided transaction data.
func TransactionFundsFlow(tx *rpcutils.Transaction) ([]*AllFundsFlows,
	[]float64, []float64, error) {
	// setLog helps avoid pushing too many log statements to the heap.
	setLog := log.Level()
	if setLog <= slog.LevelInfo {
		log.Infof("Analyzing %s data. Please Wait...", tx.TxID)
	}

	// Retrieve the inputs and outputs from the transaction's data.
	originalInputs, originalOutputs := extractAmounts(tx)
	if len(originalInputs) == 0 || len(originalOutputs) == 0 {
		return nil, originalInputs, originalOutputs,
			errors.New("funds flow check needs both input(s) and output(s) of a transaction")
	}

	if setLog <= slog.LevelInfo {
		log.Info("Generating prefabricated granular buckets from both inputs and outputs")
	}

	// granularBuckets are buckets whose inputs and outputs are split to the
	// minimum possible values.
	granularBuckets, inputs, outputs := getPrefabricatedBuckets(originalInputs, originalOutputs)

	if setLog <= slog.LevelInfo {
		log.Infof("Found %d prefabricated granular buckets from inputs and outputs",
			len(granularBuckets))
	}

	// If tx is complex exit
	if isTxComplex(inputs, outputs) {
		if setLog <= slog.LevelInfo {
			log.Infof("Complex tx %s could not be analyzed", tx.TxID)
		}

		return []*AllFundsFlows{
			&AllFundsFlows{StatusMsg: complexTxMsg},
		}, inputs, outputs, nil
	}

	if setLog <= slog.LevelInfo {
		log.Info("Calculating all possible sum combinations for both inputs and outputs")
	}

	inputCombinations := getTotalCombinations(inputs, inpointData, true)
	outputCombinations := getTotalCombinations(outputs, outpointData, true)

	// drop doping element entry if it exists.
	{
		if inputs[len(inputs)-1] == dopingElement {
			inputs = inputs[:len(inputs)-1]
		}

		if outputs[len(outputs)-1] == dopingElement {
			outputs = outputs[:len(outputs)-1]
		}
	}

	if setLog <= slog.LevelInfo {
		log.Info("Adding the outputs sums combination list to the binary tree.")
	}

	defBinaryTree := new(Node)
	if err := defBinaryTree.Insert(outputCombinations); err != nil {
		return nil, inputs, outputs,
			fmt.Errorf("Inserting the sums combinations to the binary tree failed: %v", err)
	}

	if setLog <= slog.LevelInfo {
		log.Info("Searching for matching sums between inputs and outputs amounts.")
	}

	matchedSum := defBinaryTree.FindX(inputCombinations, tx.Fees)
	if setLog <= slog.LevelInfo {
		log.Info("Matching the inputs and outputs selected to generate a solution(s)")
	}

	solutionsChan := make(chan []*AllFundsFlows)
	// getSolutions runs on a different goroutine to avoid blocking the main goroutine.
	go func() {
		sols := getSolutions(matchedSum, inputs, outputs, tx.Fees)
		solutionsChan <- sols
	}()

	txSolutions := <-solutionsChan
	close(solutionsChan)

	// ensures that matched solutions count starts from 1 always.
	for i, val := range txSolutions {
		val.Solution = i + 1
		val.FundsFlow = append(val.FundsFlow, granularBuckets...)
	}

	if setLog <= slog.LevelInfo {
		log.Infof("Found %d matching combinations solutions between the inputs and outputs",
			len(txSolutions))
	}

	// If no matching solution(s) was found then the tx is possibly resolved by default.
	if len(txSolutions) == 0 {
		txSolutions = append(txSolutions, &AllFundsFlows{
			// Solution 0, implies that the code failed to get any matches and
			// thus returned the default solution instead of null.
			Solution:  0,
			TotalFees: tx.Fees,
			FundsFlow: []TxFundsFlow{
				{
					Fee:            tx.Fees,
					Inputs:         getGroupedValues(originalInputs),
					MatchedOutputs: getGroupedValues(originalOutputs),
				},
			},
		})
	}
	return txSolutions, originalInputs, originalOutputs, nil
}

// TxFundsFlowProbability obtains the funds flow probability for each output in
// relation to its possible matching input(s).
func TxFundsFlowProbability(rawData []*AllFundsFlows,
	rawInSourceArr, rawOutSourceArr []float64) []*FlowProbability {
	if len(rawData) == 0 {
		return nil
	}

	if len(rawData) == 1 && rawData[0].StatusMsg != "" {
		return []*FlowProbability{
			&FlowProbability{StatusMsg: rawData[0].StatusMsg},
		}
	}

	// Append the amounts count to the raw source inputs slice.
	inSourceArr := appendDupsCount(rawInSourceArr)

	// Append the amounts count to the raw source outputs slice.
	outSourceArr := appendDupsCount(rawOutSourceArr)

	var totalRes []*rawResults

	log.Debug("Calculating the transaction funds flow probability...")

	allInputs := make(map[float64]int)
	// inSourceArr contains the original list of input amounts from the tx.
	for i := range inSourceArr {
		allInputs[inSourceArr[i].Amount] = inSourceArr[i].Count
	}

	allOutputs := make(map[float64]int)
	// outSourceArr contains the original list of output amount from the tx.
	for i := range outSourceArr {
		allOutputs[outSourceArr[i].Amount] = outSourceArr[i].Count
	}

	for _, entries := range rawData {
		for index := range entries.FundsFlow {
			bucket := entries.FundsFlow[index]
			g := new(rawResults)

			if g.Inputs == nil {
				g.Inputs = make(map[float64]int)
			}

			for inIndex := range bucket.Inputs.Values {
				g.Inputs[bucket.Inputs.Values[inIndex]]++
			}

			if g.MatchingOutputs == nil {
				g.MatchingOutputs = make(map[float64]*Details)
			}

			for outIndex := range bucket.MatchedOutputs.Values {
				d := bucket.MatchedOutputs.Values[outIndex]
				if g.MatchingOutputs[d] == nil {
					g.MatchingOutputs[d] = &Details{}
				}
				g.MatchingOutputs[d].Amount = bucket.MatchedOutputs.Sum
				g.MatchingOutputs[d].Count++
			}
			totalRes = append(totalRes, g)
		}
	}

	tmpRes := make(map[float64]*FlowProbability)
	for _, res := range totalRes {
		// isMany checks if the matching bucket has "many to many" or "many to
		// one" relationship between inputs and matching outputs respectively.
		// Many inputs in a bucket imply that specific output(s) cannot be easily
		// linked directly to a matching input in the same bucket as the source of
		// its funds.
		isMany := len(res.Inputs) > 1

		for out, outSum := range res.MatchingOutputs {
			// output amount to be used must be greater than zero. OP_RETURN
			// scripts do not have any amounts. They hold nulldata transaction type.
			if out <= 0 {
				continue
			}

			if tmpRes[out] == nil {
				tmpRes[out] = &FlowProbability{
					uniqueInputs: make(map[float64]int),
				}
			}

			tmpRes[out].OutputAmount = out
			tmpRes[out].Count = allOutputs[out]

			// if "many to many" or "many to one" relationship exists assign all
			// the inputs a single set.
			if isMany {
				setDetails := make([]*Details, len(res.Inputs))
				index := 0
				var isDuplicate bool
				inputsArr := make([]float64, len(res.Inputs))
				percent := roundOff((out / outSum.Amount) * float64(outSum.Count))

				for in, val := range res.Inputs {
					setDetails[index] = &Details{
						Amount:         in,
						PossibleInputs: allInputs[in],
						Actual:         val,
					}
					inputsArr[index] = in
					index++
				}

				sort.Float64s(inputsArr)

				// Check for duplicates.
				for _, set := range tmpRes[out].ProbableInputs {
					if isEqual(set.inputs, inputsArr) &&
						set.PercentOfInputs == percent {
						isDuplicate = true
						break
					}
				}

				if !isDuplicate {
					sort.Sort(byPossibleInputs(setDetails))

					tmpRes[out].ProbableInputs = append(
						tmpRes[out].ProbableInputs,
						&InputSets{Set: setDetails, PercentOfInputs: percent,
							inputs: inputsArr},
					)
				}

			} else {
				for in := range res.Inputs {
					actualCount, ok := tmpRes[out].uniqueInputs[in]
					if !ok {
						details := []*Details{&Details{
							Amount:         in,
							PossibleInputs: allInputs[in],
							Actual:         actualCount + 1,
						}}
						tmpRes[out].ProbableInputs = append(
							tmpRes[out].ProbableInputs,
							&InputSets{Set: details, PercentOfInputs: 1},
						)
						tmpRes[out].uniqueInputs[in]++
					}
				}

			}

			var rawVal float64
			for _, val := range tmpRes[out].ProbableInputs {
				if len(val.Set) != 1 {
					rawVal++
				} else {
					rawVal += float64(val.Set[0].PossibleInputs)
				}
			}

			tmpRes[out].LinkingProbability = roundOff(1 / rawVal)
		}
	}

	log.Debug("Finished calculating the tx probabilities.")

	var data []*FlowProbability
	for _, ss := range tmpRes {
		data = append(data, ss)
	}

	return data
}
