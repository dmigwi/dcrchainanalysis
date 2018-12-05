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
	Inputs          map[float64]int
	MatchingOutputs map[float64]*Details
}

// Details defines the input or output amount value and its duplicates count.
type Details struct {
	Amount float64
	Count  int
}

// InputSets groups probable inputs into sets each with its own percent of input value.
type InputSets struct {
	Set             []*Details
	PercentOfInputs float64
	Inputs          []float64
}

// FlowProbability defines the final transaction funds flow data that includes
// the output tx funds flow probability.
type FlowProbability struct {
	OutputAmount       float64
	Count              int
	LinkingProbability float64
	ProbableInputs     []*InputSets
	uniqueInputs       map[float64]int
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

	log.Debugf("The transaction has %d inputs and %d outputs amounts repectively",
		len(inputs), len(outputs))

	return
}

// TransactionFundsFlow calculates the funds flow between a set of inputs and
// their corresponding set of outputs for the provided transaction data.
func TransactionFundsFlow(tx *datatypes.Transaction) ([]*AllFundsFlows, []float64, []float64, error) {
	// Retrieve the inputs and outputs from the transaction's data.
	inputs, outputs := extractAmounts(tx)
	if len(inputs) == 0 || len(outputs) == 0 {
		return nil, inputs, outputs,
			errors.New("funds flow check needs both input(s) and output(s) of a transaction")
	}

	log.Info("Calculating all possible sum combinations for both inputs and outputs")

	inputCombinations := getTotalCombinations(inputs, inpointData)
	outputCombinations := getTotalCombinations(outputs, outpointData)

	log.Info("Adding the outputs sums combination list to the binary tree.")

	defBinaryTree := new(Node)
	if err := defBinaryTree.Insert(outputCombinations); err != nil {
		return nil, inputs, outputs,
			fmt.Errorf("Inserting the sums combinations to the bianry tree failed: %v", err)
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
				// split the funds flow buckets into their most granular buckets.
				tmp = splitFundsFlow(tmp)

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

	return sol[maxBucketsCount], inputs, outputs, nil
}

// equals works effectively when the inputs and output combinations are sorted.
func (f *AllFundsFlows) equals(item *AllFundsFlows) bool {
	var matchedBuckets, totalBuckets int

	for _, bucket := range item.FundsFlow {
		var isInMatch, isOutMatch bool
		for _, elem := range f.FundsFlow {
			if reflect.DeepEqual(bucket.Inputs.Values,
				elem.Inputs.Values) {
				isInMatch = true
			}

			if reflect.DeepEqual(bucket.MatchedOutputs.Values,
				elem.MatchedOutputs.Values) {
				isOutMatch = true
			}
		}
		if isInMatch && isOutMatch {
			matchedBuckets++
		}
		totalBuckets++
	}

	if totalBuckets == matchedBuckets {
		return true
	}
	return false
}

// splitFundsFlow breaks down the buckets into their most granular form using one
// of the buckets which is a duplicate in the combined bucket. Because of
// computational power limitations the GenerateCombinations doesn't produces
// duplicate combinations unless the combination length r is 1. Since possible
// combinations are greatly reduced by the time this function is invoked its
// safer to do the bucket spliting here than in the GenerateCombinations function.
func splitFundsFlow(combined []*TxFundsFlow) []*TxFundsFlow {
	var newData = make([]*TxFundsFlow, 0)
	for ind2 := 0; ind2 < len(combined); ind2++ {
		b2 := combined[ind2]

		for ind1 := 0; ind1 < len(combined); ind1++ {
			b1 := combined[ind1]
			if (len(b2.Inputs.Values) > len(b1.Inputs.Values)) &&
				(len(b2.MatchedOutputs.Values) > len(b1.MatchedOutputs.Values)) {
				inputsDiff, inputSum := arrayDiff(
					b1.Inputs.Values, b2.Inputs.Values)
				outputDiff, outputSum := arrayDiff(
					b1.MatchedOutputs.Values, b2.MatchedOutputs.Values)

				if roundOff(inputSum-outputSum+b1.Fee) == b2.Fee {
					newData = append(newData, b1, &TxFundsFlow{
						Fee:            roundOff(inputSum - outputSum),
						Inputs:         &GroupedValues{Sum: inputSum, Values: inputsDiff},
						MatchedOutputs: &GroupedValues{Sum: outputSum, Values: outputDiff},
					})

					combined = append(combined[:ind2], combined[ind2+1:]...)
					break
				}
			}
		}
	}
	if len(newData) > 0 {
		return append(combined, newData...)
	}
	return combined
}

// arrayDiff returns the difference between arr2 and arr1 i.e. arr2 - arr1.
func arrayDiff(arr1, arr2 []float64) (tmp []float64, sum float64) {
	tmp = make([]float64, len(arr2))
	copy(tmp, arr2)

	for _, val := range arr1 {
		for i, val2 := range tmp {
			if val == val2 && i < len(tmp) {
				tmp = append(tmp[:i], tmp[i+1:]...)
				break
			}
		}
	}
	for _, entry := range tmp {
		sum += entry
	}
	return tmp, roundOff(sum)
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
func TxFundsFlowProbability(rawData []*AllFundsFlows,
	inSourceArr, outSourceArr []float64) []*FlowProbability {
	totalRes := make([]*RawResults, 0)
	if len(rawData) == 0 {
		return nil
	}

	log.Debug("Calculating the transaction funds flow probabiblity...")

	allInputs := make(map[float64]int, 0)
	// inSourceArr contains the original list of input amounts from the tx.
	for _, val := range inSourceArr {
		allInputs[val]++
	}

	allOutputs := make(map[float64]int, 0)
	// outSourceArr contains the original list of output amount from the tx.
	for _, val := range outSourceArr {
		allOutputs[val]++
	}

	for _, entries := range rawData {
		for _, bucket := range entries.FundsFlow {
			g := new(RawResults)

			if g.Inputs == nil {
				g.Inputs = make(map[float64]int, 0)
			}

			for _, a := range bucket.Inputs.Values {
				g.Inputs[a]++
			}

			if g.MatchingOutputs == nil {
				g.MatchingOutputs = make(map[float64]*Details, 0)
			}

			for _, d := range bucket.MatchedOutputs.Values {
				if g.MatchingOutputs[d] == nil {
					g.MatchingOutputs[d] = &Details{}
				}
				g.MatchingOutputs[d].Amount = bucket.MatchedOutputs.Sum
				g.MatchingOutputs[d].Count++
			}
			totalRes = append(totalRes, g)
		}
	}

	tmpRes := make(map[float64]*FlowProbability, 0)
	for _, res := range totalRes {
		// isManyToMany checks if the matching bucket has many to many or many to
		// one relationship between inputs and matching outputs. Many inputs in
		// bucket means that specific output(s) cannot be easily linked directly
		// to matching input in the same bucket as the source of funds.
		isManyToMany := len(res.Inputs) > 1

		for out, outSum := range res.MatchingOutputs {
			if tmpRes[out] == nil {
				tmpRes[out] = &FlowProbability{
					uniqueInputs: make(map[float64]int, 0),
				}
			}

			tmpRes[out].OutputAmount = out
			tmpRes[out].Count = allOutputs[out]

			// if Many to many relationship exists assign all the inputs a single set.
			if isManyToMany {
				setDetails := make([]*Details, len(res.Inputs))
				index := 0
				var isDuplicate bool
				inputsArr := make([]float64, len(res.Inputs))
				percent := math.Round((out/outSum.Amount)*1000000) / 10000 * float64(outSum.Count)

				for in := range res.Inputs {
					setDetails[index] = &Details{Amount: in, Count: allInputs[in]}
					inputsArr[index] = in
					index++
				}

				sort.Float64s(inputsArr)

				// Check for duplicates
				for _, set := range tmpRes[out].ProbableInputs {
					if reflect.DeepEqual(set.Inputs, inputsArr) &&
						set.PercentOfInputs == roundOff(percent) {
						isDuplicate = true
						break
					}
				}

				if !isDuplicate {
					tmpRes[out].ProbableInputs = append(
						tmpRes[out].ProbableInputs,
						&InputSets{Set: setDetails, PercentOfInputs: roundOff(percent),
							Inputs: inputsArr},
					)
				}

			} else {
				for in := range res.Inputs {
					_, ok := tmpRes[out].uniqueInputs[in]
					if !ok {
						details := []*Details{&Details{Amount: in, Count: allInputs[in]}}
						tmpRes[out].ProbableInputs = append(
							tmpRes[out].ProbableInputs,
							&InputSets{Set: details, PercentOfInputs: 100},
						)

						tmpRes[out].uniqueInputs[in]++
					}
				}

			}
			rawVal := float64(len(tmpRes[out].ProbableInputs))
			tmpRes[out].LinkingProbability = math.Round((1/rawVal)*10000) / 100
		}
	}

	log.Debug("Finished calculating the tx probabilities.")

	var data []*FlowProbability
	for _, ss := range tmpRes {
		data = append(data, ss)
	}

	return data
}
