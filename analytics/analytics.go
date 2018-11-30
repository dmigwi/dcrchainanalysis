// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package analytics

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"

	"github.com/raedahgroup/dcrchainanalysis/v1/datatypes"
)

type txProperties string

const (
	// inpointData defines the input type of data.
	inpointData txProperties = "inputs"

	// outpointData defines the ouput type of data.
	outpointData txProperties = "outputs"
)

// TxFundsFlow link inputs with their matching outputs.
type TxFundsFlow struct {
	Fee            float64
	Inputs         *GroupedValues
	MatchedOutputs *GroupedValues
}

// AllFundsFlows groups together all the possible solutions to the inputs and
// outputs funds flow.
type AllFundsFlows struct {
	Solution  int
	TotalFees float64
	FundsFlow []*TxFundsFlow
}

// RawResults defines some compressed solutions data needed for further processing
// of the transaction funds flow.
type RawResults struct {
	Inputs          []float64
	MatchingOutputs map[float64]int
}

// FlowProbability defines the final transaction funds flow data that includes
// the output tx funds flow probability.
type FlowProbability struct {
	OutputAmount   float64
	ProbableInputs []float64
	uniqueInputs   map[float64]float64
	Probability    float64
}

// extractAmounts retrieves the transaction input(s) and output(s) and returns
// sorted slices are sort.
func extractAmounts(data *datatypes.Transaction) (inputs, outputs []float64) {
	inputs = make([]float64, data.NumInpoint, data.NumInpoint)
	for i, entry := range data.Inpoints {
		inputs[i] = entry.ValueIn
	}

	outputs = make([]float64, data.NumOutpoint, data.NumOutpoint)
	for i, entry := range data.Outpoints {
		outputs[i] = entry.Value
	}

	sort.Float64s(inputs)
	sort.Float64s(outputs)

	return
}

// TransactionFundsFlow calculates the funds flow between a set of inputs and
// their corresponding set of outputs for the provided transaction data.
func TransactionFundsFlow(tx *datatypes.Transaction) ([]*AllFundsFlows, error) {
	// Retrieve the inputs and outputs from the transaction's data.
	inputs, outputs := extractAmounts(tx)
	if len(inputs) == 0 || len(outputs) == 0 {
		return nil, errors.New("funds flow check needs both input(s) and output(s) of a transaction")
	}

	log.Info("Calculating all possible sum combinations for both inputs and outputs")

	inputCombinations := getTotalCombinations(inputs, inpointData)
	outputCombinations := getTotalCombinations(outputs, outpointData)

	log.Info("Adding the outputs sums combination list to the binary tree.")

	defBinaryTree := new(Node)
	if err := defBinaryTree.Insert(outputCombinations); err != nil {
		return nil, fmt.Errorf("Inserting the sums combinations to the bianry tree failed: %v", err)
	}

	log.Info("Searching for matching sums between inputs and outputs amounts.")
	var matchedSum []*TxFundsFlow

	for _, in := range inputCombinations {
		marchedArr := defBinaryTree.FindX(in, tx.Fees)
		if len(marchedArr) != 0 {
			matchedSum = append(matchedSum, marchedArr...)
		}
	}

	var maxBucketsCount int
	count := 1 // Solutions count
	target := tx.Fees
	sol := make(map[int][]*AllFundsFlows, 0)

	log.Trace("Matching the inputs and outputs selected to generate a solution")

	for index := range matchedSum {
		var sumFees float64
		var tmp []*TxFundsFlow

		inputCopy := make([]float64, len(inputs), len(inputs))
		outputCopy := make([]float64, len(outputs), len(outputs))

		copy(inputCopy, inputs)
		copy(outputCopy, outputs)
		log.Trace(" \n ")

		// Reorder the matchedSum slice by changing its start to end value
		// while maintaining the original slice items following order. A linked
		// slice is created and versions count equal to its length are created.
		for k, val := range append(matchedSum[index:], matchedSum[:index]...) {
			log.Tracef("###### index: %d ###### ToTalFee: %f ###### Fee: %f", k, sumFees, target)

			if val.Fee <= roundOff(target-sumFees) {
				var totalIn, totalOut int

				inCopy := make([]float64, len(inputCopy), len(inputCopy))
				outCopy := make([]float64, len(outputCopy), len(outputCopy))

				copy(inCopy, inputCopy)
				copy(outCopy, outputCopy)

				log.Tracef(" Possible new solution with input Values: %v and output Values: %v",
					val.Inputs.Values, val.MatchedOutputs.Values)

				log.Tracef(" Initial status of inputCopy: %v and outputCopy: %v",
					inputCopy, outputCopy)

				for _, entry := range val.Inputs.Values {
				inputCopyLoop:
					for i, in := range inputCopy {
						if entry == in {
							inputCopy = append(inputCopy[:i], inputCopy[i+1:]...)
							totalIn++
							break inputCopyLoop
						}
					}
				}

				log.Tracef(" Modified inputCopy: %v", inputCopy)

				// If all the inputs were not in the inputsCopy array restore
				// inputCopy to their earlier version. Only inputCopy that has
				// modified.
				if totalIn != len(val.Inputs.Values) {
					inputCopy = make([]float64, len(inCopy), len(inCopy))

					copy(inputCopy, inCopy)

					log.Tracef(" Restored inputCopy to %v since %d entry(s) did not match",
						inputCopy, len(val.Inputs.Values)-totalIn)

					goto ifStatementEnd
				}

				for _, entry := range val.MatchedOutputs.Values {
				outCopyLoop:
					for i, out := range outputCopy {
						if entry == out {
							outputCopy = append(outputCopy[:i], outputCopy[i+1:]...)
							totalOut++
							break outCopyLoop
						}
					}
				}

				log.Tracef(" Modified outputCopy: %v", outputCopy)

				// If all the outputs were not in the outputCopy array restore
				// inputCopy and outputCopy to their earlier version.
				if totalOut != len(val.MatchedOutputs.Values) {
					inputCopy = make([]float64, len(inCopy), len(inCopy))
					outputCopy = make([]float64, len(outCopy), len(outCopy))

					copy(inputCopy, inCopy)
					copy(outputCopy, outCopy)

					log.Tracef(" Restored inputCopy to %v and outputCopy to %v since %d "+
						"and %d entries did not match respectively", inputCopy, outputCopy,
						len(val.Inputs.Values)-totalIn, len(val.MatchedOutputs.Values)-totalOut)

					goto ifStatementEnd
				}

				sumFees += val.Fee

				log.Tracef(" Matched part of solution with input values: %v and"+
					" output values: %v was selected", val.Inputs.Values,
					val.MatchedOutputs.Values)

				tmp = append(tmp, &TxFundsFlow{
					Fee:            roundOff(val.Fee),
					Inputs:         val.Inputs,
					MatchedOutputs: val.MatchedOutputs,
				})
			ifStatementEnd:
			}

			// append all the matched solutions
			if sumFees/target > 0.99 && len(inputCopy) == 0 && len(outputCopy) == 0 {
				// If current solution has too few buckets ignore it.
				if len(tmp) >= maxBucketsCount {
					maxBucketsCount = len(tmp)

					item := &AllFundsFlows{
						Solution:  count,
						TotalFees: roundOff(sumFees),
						FundsFlow: tmp,
					}

					var isDuplicate bool
					// Check for duplicates
				dupsLoop:
					for _, elem := range sol {
						for _, s := range elem {
							isDuplicate = s.equals(item)
							if isDuplicate {
								break dupsLoop
							}
						}
					}

					if !isDuplicate {
						sol[len(tmp)] = append(sol[len(tmp)][:], item)
						count++
					}
				}
				sumFees = 0.0
			}

			// No input and output matching that will happen if either is empty
			if len(inputCopy) == 0 || len(outputCopy) == 0 {
				break
			}
		}
	}

	log.Infof("Found %d matching sums between the inputs and outputs",
		len(sol[maxBucketsCount]))

	return sol[maxBucketsCount], nil
}

// equals works effectively when the inputs and output combinations are sorted
func (f *AllFundsFlows) equals(item *AllFundsFlows) bool {
	var inputsCount, outputsCount int
	for _, elem := range f.FundsFlow {
		for _, bucket := range item.FundsFlow {
			if reflect.DeepEqual(bucket.Inputs.Values, elem.Inputs.Values) {
				inputsCount++
			}

			if reflect.DeepEqual(bucket.MatchedOutputs.Values, elem.MatchedOutputs.Values) {
				outputsCount++
			}
		}
	}
	if inputsCount == len(item.FundsFlow) && outputsCount == len(item.FundsFlow) {
		return true
	}
	return false
}

// getTotalCombinations fetches all the possible combinations of the source
// array except when the elements of the combinations (its length) is equal to the
// source array length.
func getTotalCombinations(sourceArr []float64, p txProperties) (totalCombinations []*GroupedValues) {
	log.Infof("Calculating %s set sum amount combinations.", p)

	for i := 1; i < len(sourceArr); i++ {
		combinations := GenerateCombinations(sourceArr, int64(i))
		totalCombinations = append(totalCombinations, combinations...)
	}

	log.Debugf("Found %d %s possible sum combinations", len(totalCombinations), p)
	return
}

// TxFundsFlowProbability obtains the funds flow probability for each output in
// relation to its possible matching input(s).
func TxFundsFlowProbability(rawData []*AllFundsFlows) []*FlowProbability {
	totalRes := make([]*RawResults, 0)
	if len(rawData) == 0 {
		return nil
	}

	log.Debug("Calculating the transaction funds flow probabiblity...")

	for _, entries := range rawData {
		for _, bucket := range entries.FundsFlow {
			g := &RawResults{Inputs: bucket.Inputs.Values}

			if g.MatchingOutputs == nil {
				g.MatchingOutputs = make(map[float64]int, 0)
			}

			for _, d := range bucket.MatchedOutputs.Values {
				g.MatchingOutputs[d]++
			}
			totalRes = append(totalRes, g)
		}
	}

	tmpRes := make(map[float64]*FlowProbability, 0)
	for _, tt := range totalRes {
		for outVal := range tt.MatchingOutputs {
			if tmpRes[outVal] == nil {
				tmpRes[outVal] = &FlowProbability{
					uniqueInputs: make(map[float64]float64, 0),
				}
			}

			tmpRes[outVal].OutputAmount = outVal
			for _, inVal := range tt.Inputs {
				_, ok := tmpRes[outVal].uniqueInputs[inVal]
				if !ok && inVal >= outVal {
					tmpRes[outVal].ProbableInputs = append(
						tmpRes[outVal].ProbableInputs, inVal,
					)

					tmpRes[outVal].uniqueInputs[inVal] = inVal
					tmpRes[outVal].Probability = math.Round((100/float64(
						len(tmpRes[outVal].ProbableInputs)))*100) / 100
				}
			}
		}
	}

	log.Debug("Finished calculating the tx probabilities.")

	var data []*FlowProbability
	for _, ss := range tmpRes {
		data = append(data, ss)
	}

	return data
}
