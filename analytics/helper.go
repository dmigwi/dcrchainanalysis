// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package analytics

import (
	"math"
)

const (
	// dopingElement is a placeholder value that helps guarrantee accuracy in
	// generating sum combinations with no duplicates when the source slice has
	// its last element as a duplicate.
	dopingElement float64 = -1
)

// GroupedValues clusters together values as duplicates or other grouped values.
// It holds the total sum and the list of the duplicates/grouped values.
type GroupedValues struct {
	Sum    float64   `json:",omitempty"`
	Values []float64 `json:",omitempty"`
}

// GeneratePermutations calculates the count of all possible outcomes with n as
// the whole set of data and r as the expected subset of the various elements to
// form a single value.
func GeneratePermutations(n, r int64) int64 {
	return permutation(n, n-r)
}

// permutation is the recusive function that generates the actual permutations.
// It can also be used to calculate any factorial by just setting r as 1.
func permutation(n, r int64) int64 {
	if r < 1 {
		r = 1
	}
	if n == r || n <= 1 {
		return 1
	}
	return permutation(n-1, r) * n
}

// GroupDuplicates uses a map to group together duplicates where its keys are
// unique and its value should list all the duplicates.
func GroupDuplicates(list []float64) map[float64]*GroupedValues {
	d := make(map[float64]*GroupedValues, 0)
	sum := 0.0

	for ind := range list {
		s := d[list[ind]]
		if s == nil {
			s = &GroupedValues{}
		}
		sum = s.Sum
		sum += list[ind]
		s.Sum = roundOff(sum)
		s.Values = append(s.Values, list[ind])
		d[list[ind]] = s
	}
	return d
}

// appendDupsCount adds a duplicates count value to each element is the array
func appendDupsCount(list []float64) (details []*Details) {
	gD := GroupDuplicates(list)
	details = make([]*Details, len(list))

	for index := range list {
		val := list[index]
		details[index] = &Details{
			Amount: val,
			Count:  len(gD[val].Values),
		}
	}

	// Add the doping element when the last entry in the slice is duplicate.
	if details[len(list)-1].Count > 1 {
		details = append(details, &Details{Amount: dopingElement, Count: 1})
	}

	return
}

// GenerateCombinations generates all the combinations for the array with the
// subset r count provided.
func GenerateCombinations(sourceArray []*Details, r int64) []*GroupedValues {
	output := make(chan []float64)

	go func(newSource []*Details, rVal int64, outputChan chan<- []float64) {
		var newArrayIndex, oldArrayIndex int64
		data := make([]float64, r)

		combinatorics(newSource, rVal, newArrayIndex, oldArrayIndex, data, outputChan)
		close(output)
	}(sourceArray, r, output)

	result := make([]*GroupedValues, 0)

	for elem := range output {
		var sum float64
		for i := range elem {
			sum += elem[i]
		}
		result = append(result, &GroupedValues{
			Values: elem,
			Sum:    roundOff(sum),
		})
	}

	return result
}

// combinatorics is a recusive function that generates all the combinations C of
// subset r values from a set of n values. i.e nCr = n-1 C r-1 + n-1 C
func combinatorics(source []*Details, r, newArrInd, sourceArrInd int64,
	data []float64, output chan<- []float64) {
	if newArrInd == r && data[r-1] != dopingElement {
		var tmp = make([]float64, r)
		copy(tmp, data)
		output <- tmp
		return
	}

	// value of sourceArrInd should always be less than the length of source
	// array to avoid 'index out of range' errors with source slice.
	if sourceArrInd >= int64(len(source)) {
		return
	}

	// when r = 1 keep all the duplicates
	for i := 0; r > 1; {
		if data[newArrInd] == source[sourceArrInd].Amount {
			v := sourceArrInd + 1
			if v < int64(len(source)) {
				sourceArrInd = v
			}
		}

		if i >= source[sourceArrInd].Count {
			break
		}

		i++
	}

	data[newArrInd] = source[sourceArrInd].Amount

	combinatorics(source, r, newArrInd+1, sourceArrInd+1, data, output)
	combinatorics(source, r, newArrInd, sourceArrInd+1, data, output)
}

// ExtractSums retrieves all the sum values from the input map into a slice
// after all the duplicates were summed up together.
func ExtractSums(val map[float64]*GroupedValues) ([]float64, map[float64][]float64) {
	sums := make([]float64, len(val), len(val))
	sumsToUniqueVals := make(map[float64][]float64, len(val))
	i := 0
	for _, val := range val {
		sums[i] = val.Sum
		sumsToUniqueVals[val.Sum] = val.Values
		i++
	}

	return sums, sumsToUniqueVals
}

// rounds off the float value to a value with eight decimals places.
func roundOff(item float64) float64 {
	return math.Round(item*10e8) / 10e8
}
