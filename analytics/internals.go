// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package analytics

import (
	"sort"

	"github.com/raedahgroup/dcrchainanalysis/v1/rpcutils"
)

// extractAmounts retrieves the transaction input(s) and output(s) and returns
// sorted slices. It appends the amount count to the source slice. If the last
// element in the sort slice is duplicate the dopingElement is appended.
func extractAmounts(data *rpcutils.Transaction) (inputs, outputs []float64) {
	for i := range data.Inpoints {
		inputs = append(inputs, data.Inpoints[i].ValueIn)
	}

	for i := range data.Outpoints {
		outputs = append(outputs, data.Outpoints[i].Value)
	}

	sort.Float64s(inputs)
	sort.Float64s(outputs)

	// Add the doping element when the last entry in the slice is a duplicate.
	if len(inputs) > 1 && inputs[len(inputs)-1] == inputs[len(inputs)-2] {
		inputs = append(inputs, dopingElement)
	}

	if len(outputs) > 1 && outputs[len(outputs)-1] == outputs[len(outputs)-2] {
		outputs = append(outputs, dopingElement)
	}

	log.Debugf("The transaction has %d inputs and %d outputs amounts respectively",
		len(inputs), len(outputs))

	return
}

// getTotalCombinations fetches all the possible combinations of the source
// array except when the elements of the combinations (its length) is equal to the
// source array length.
func getTotalCombinations(sourceArr []float64, p txProperties, isLog ...bool) (
	totalCombinations []GroupedValues) {
	if len(isLog) > 0 && isLog[0] {
		log.Infof("Calculating %s set sum amount combinations.", p)
	}

	// Start calculating the largest combinations.
	for i := int64(len(sourceArr) - 1); i > 0; i-- {
		totalCombinations = append(totalCombinations, GenerateCombinations(sourceArr, i)...)
	}

	if len(isLog) > 0 && isLog[0] {
		log.Debugf("Found %d %s possible sum combinations", len(totalCombinations), p)
	}
	return
}

// Using the txfee getSolutions returns the most detailed solution(s) generated
// by ensuring that every solution has a total txfee equivalent to the original
// txfee value. It also ensures that individual count of all inputs and outputs
// match what is in the original transaction data.
func getSolutions(data []TxFundsFlow, inputs, outputs []float64,
	txFees float64) []*AllFundsFlows {
	var maxBuckets, index int

	matchedSumCopy := make([]TxFundsFlow, len(data))

	inputCopy := make([]float64, len(inputs))
	outputCopy := make([]float64, len(outputs))
	inCopy := make([]float64, len(inputs))
	outCopy := make([]float64, len(outputs))

	temp := make(map[int][]*AllFundsFlows)

	for index = range data {
		// use the descending order to rearrange the slice because the needed results
		// are mostly at the end of the matchedSum array.
		index = len(data) - index

		copy(matchedSumCopy[:len(data[index:])], data[index:])
		copy(matchedSumCopy[len(data[index:]):], data[:index])

		copy(inputCopy, inputs)
		copy(outputCopy, outputs)

		getSolutionsWorker(inputCopy, outputCopy, inCopy, outCopy,
			matchedSumCopy, &temp, &maxBuckets, txFees)
	}

	return temp[maxBuckets]
}

// getSolutionsWorker obtains a single solution from a given arrangement of the
// matched inputs and outputs.
func getSolutionsWorker(inputCopy, outputCopy, inCopy, outCopy []float64,
	matchedSumCopy []TxFundsFlow, payload *map[int][]*AllFundsFlows,
	maxBucketsCount *int, txFees float64) {
	var tmp []TxFundsFlow
	var sumFees float64
	var inSwapIndex, outSwapIndex int
	var restoreInputs, restoreOutputs, isDuplicate bool

	// Reorder the matchedSum slice by changing its start to end value
	// while maintaining the original slice items following order. A linked
	// slice is created and versions count equal to its length are created.
	for k := range matchedSumCopy {
		val := matchedSumCopy[k]

		if val.Fee <= roundOff(txFees-sumFees) {
			restoreInputs, restoreOutputs = false, false

			inCopy = inCopy[:len(inputCopy)]
			copy(inCopy, inputCopy)

			outCopy = outCopy[:len(outputCopy)]
			copy(outCopy, outputCopy)

			inSwapIndex, restoreInputs = compareBucketIO(val.Inputs.Values, inputCopy)

			// do not process outputs if inputs are to be restored.
			if !restoreInputs {
				outSwapIndex, restoreOutputs = compareBucketIO(val.MatchedOutputs.Values, outputCopy)
			}

			if !restoreInputs && !restoreOutputs {
				inputCopy = inputCopy[:inSwapIndex]
				outputCopy = outputCopy[:outSwapIndex]

				sumFees += val.Fee
				tmp = append(tmp, TxFundsFlow{
					Fee:            roundOff(val.Fee),
					Inputs:         val.Inputs,
					MatchedOutputs: val.MatchedOutputs,
				})
			}

			switch {
			case restoreOutputs:
				copy(outputCopy, outCopy)
				fallthrough

			case restoreInputs:
				copy(inputCopy, inCopy)
			}
		}

		// roundoff to 8 decimal places
		sumFees = roundOff(sumFees)

		// append all the matched solutions
		if sumFees == txFees && len(inputCopy) == 0 && len(outputCopy) == 0 {
			// split the funds flow buckets into their most granular buckets.
			tmp = splitFundsFlow(tmp)

			// If current solution has too few buckets ignore it.
			if len(tmp) >= *maxBucketsCount {
				item := AllFundsFlows{
					TotalFees: sumFees,
					FundsFlow: tmp,
				}

				// Check for duplicates
			dupsLoop:
				for _, elem := range *payload {
					for _, s := range elem {
						isDuplicate = s.equals(item)
						if isDuplicate {
							break dupsLoop
						}
					}
				}

				*maxBucketsCount = len(tmp)

				if !isDuplicate {
					(*payload)[len(tmp)] = append((*payload)[len(tmp)][:], &item)
				}
			}
		}

		// No input and output matching that will happen if either is empty
		if len(inputCopy) == 0 || len(outputCopy) == 0 {
			return
		}
	}
}

// compareBucketIO checks if the given inputs/outputs from the current bucket
// exists in the copy of the original inputs/outputs respectively. This is used to
// assertain that all possible tx solutions have all the inputs and outputs values
// accounted for. Once a given bucketEntry amount entry is matched to another
// amount in the original copy, the matched amount in the original copy is moved
// to the back of the slice, a lastSwapIndex keep advancing towards the start of
// the slice to indicate where the valid/unmatched array items end at.
func compareBucketIO(bucketEntry, origCopy []float64) (int, bool) {
	var count int
	var lastSwapIndex = len(origCopy)

	for _, entry := range bucketEntry {
		origCopy = origCopy[:lastSwapIndex]

		for i := range origCopy {
			if entry == origCopy[i] {
				lastSwapIndex--

				origCopy[i] = origCopy[lastSwapIndex]

				count++
				break
			}
		}
	}
	return lastSwapIndex, len(bucketEntry) != count
}

// splitFundsFlow breaks down the buckets into their most granular form. Because of
// computational power limitations GenerateCombinations doesn't produce granular
// duplicate combinations unless the combination length r is 1. Since possible
// combinations are greatly reduced by the time this function is invoked, its
// cheaper to do the bucket spliting here than in GenerateCombinations function.
func splitFundsFlow(combined []TxFundsFlow) []TxFundsFlow {
	var combinations []GroupedValues

mainLoop:
	for i := 0; i < len(combined); i++ {
		f := combined[i]
		if len(f.Inputs.Values) > 1 && len(f.MatchedOutputs.Values) > 1 {
			combinations = getTotalCombinations(f.Inputs.Values, inpointData)

			for k := 1; k < len(f.MatchedOutputs.Values); k++ {
				var newArrInd, sourceArrInd int64
				var results []GroupedValues
				var data = make([]float64, k)

				combinatorics(&results, f.MatchedOutputs.Values, int64(k),
					newArrInd, sourceArrInd, data)

				for m := range combinations {
					for n := range results {
						diff := combinations[m].Sum - results[n].Sum
						if diff >= 0 && diff < f.Fee {
							combined[i].Fee = roundOff(diff)
							combined[i].Inputs = combinations[m]
							combined[i].MatchedOutputs = results[n]

							diffOut, SumOut := arrayDiff(results[n].Values,
								f.MatchedOutputs.Values)
							diffIn, SumIn := arrayDiff(combinations[m].Values,
								f.Inputs.Values)

							combined = append(combined, TxFundsFlow{
								Fee: roundOff(SumIn - SumOut),
								Inputs: GroupedValues{Sum: SumIn,
									Values: diffIn},
								MatchedOutputs: GroupedValues{Sum: SumOut,
									Values: diffOut},
							})
							i = 0
							continue mainLoop
						}
					}
				}
			}
		}
	}
	return combined
}

// arrayDiff returns the difference between arr2 and arr1 i.e. arr2 - arr1.
func arrayDiff(arr1, arr2 []float64) (tmp []float64, sum float64) {
	tmp = make([]float64, len(arr2))
	copy(tmp, arr2)

	for k := range arr1 {
		for i := range tmp {
			if arr1[k] == tmp[i] && i < len(tmp) {
				copy(tmp[i:], tmp[i+1:])
				tmp = tmp[:len(tmp)-1]
				break
			}
			// No more possible match exist at this case scenario
			if tmp[i] > arr1[k] {
				break
			}
		}
	}

	// if a subset of arr1 was not found in arr2 return empty tmp array and
	// zero sum value.
	if len(tmp) == len(arr2) {
		tmp = tmp[:0]
		return
	}

	for _, entry := range tmp {
		sum += entry
	}

	return tmp, roundOff(sum)
}

// equals works effectively when the inputs and output combinations are sorted.
// It checks the equality of two solutions i.e. returns true if the current
// solution matches an earlier generated solution.
func (f *AllFundsFlows) equals(item AllFundsFlows) bool {
	var matchedBuckets, totalBuckets int

	for _, bucket := range item.FundsFlow {
		var isInMatch, isOutMatch bool
		for _, elem := range f.FundsFlow {
			if isEqual(bucket.Inputs.Values, elem.Inputs.Values) {
				isInMatch = true
			}

			if isEqual(bucket.MatchedOutputs.Values, elem.MatchedOutputs.Values) {
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

// getPrefabricatedBuckets helps reduce the number of inputs and outputs to be
// processed via the combinations method to generate possible buckets with matching
// inputs and outputs. It creates a slice of buckets whose inputs can be easily
// matched to specific outputs.
func getPrefabricatedBuckets(inputs, outputs []float64) (
	solutions []TxFundsFlow, newInputs, newOutputs []float64) {
	newInputs = make([]float64, len(inputs))
	copy(newInputs, inputs)

	newOutputs = make([]float64, len(outputs))
	copy(newOutputs, outputs)

	for in := 0; in < len(newInputs); in++ {
	subLoop:
		for out := 0; out < len(newOutputs); out++ {
			// if a specific inputs can be matched to a specific output then
			// that is one of the granular buckets needed to generate the final
			// solution.
			if newInputs[in] == newOutputs[out] && newInputs[in] != dopingElement {
				solutions = append(solutions, TxFundsFlow{
					Fee:            0,
					Inputs:         getGroupedValues(newInputs[in : in+1]),
					MatchedOutputs: getGroupedValues(newOutputs[out : out+1]),
				})

				copy(newInputs[in:], newInputs[in+1:])
				newInputs = newInputs[:len(newInputs)-1]

				copy(newOutputs[out:], newOutputs[out+1:])
				newOutputs = newOutputs[:len(newOutputs)-1]

				in = 0
				break subLoop
			}
		}
	}

	return
}
