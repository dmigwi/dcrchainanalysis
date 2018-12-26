// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package analytics

import (
	"encoding/json"
	"math"
)

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
func GroupDuplicates(list []float64) map[float64]GroupedValues {
	d := make(map[float64]GroupedValues)
	sum := 0.0

	for ind := range list {
		s := d[list[ind]]
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

	for _, val := range list {
		details = append(details, &Details{
			Amount: val,
			Count:  len(gD[val].Values),
		})
	}

	return
}

// GenerateCombinations generates all the combinations for the array with the
// subset r count provided.
func GenerateCombinations(sourceArray []float64, r int64) []GroupedValues {
	output := make(chan []GroupedValues)
	defer close(output)

	go func(newSource []float64, rVal int64, outputChan chan<- []GroupedValues) {
		var newArrayIndex, oldArrayIndex int64
		var res []GroupedValues
		data := make([]float64, r)

		combinatorics(&res, newSource, rVal, newArrayIndex, oldArrayIndex, data)
		outputChan <- res
	}(sourceArray, r, output)

	return <-output
}

// combinatorics is a recusive function that generates all the combinations C of
// subset r values from a set of n values. i.e nCr = n-1 C r-1 + n-1 C
func combinatorics(res *[]GroupedValues, source []float64, r, newArrInd, sourceArrInd int64,
	data []float64) {
	if newArrInd == r && data[r-1] != dopingElement {
		*res = append(*res, getGroupedValues(data))
		return
	}

	// value of sourceArrInd should always be less than the length of source
	// array to avoid 'index out of range' errors with source slice.
	if sourceArrInd >= int64(len(source)) {
		return
	}

	// when r = 1 keep all the duplicates else jump till end of duplicates.
	for r > 1 {
		if data[newArrInd] == source[sourceArrInd] &&
			(sourceArrInd+1) < int64(len(source)) {
			sourceArrInd++
		} else {
			break
		}

		if sourceArrInd == int64(len(source)) ||
			source[sourceArrInd-1] != source[sourceArrInd] {
			break
		}
	}

	data[newArrInd] = source[sourceArrInd]

	combinatorics(res, source, r, newArrInd+1, sourceArrInd+1, data)
	combinatorics(res, source, r, newArrInd, sourceArrInd+1, data)
}

// rounds off the float value to a value with eight decimals places.
func roundOff(item float64) float64 {
	return math.Round(item*10e8) / 10e8
}

// isEqual is a slice equality check function. reflect.DeepEquals is the most
// accurate when checking the slice equality but has a lot of overheads which
// includes but not limited to making unnecessary allocations.
func isEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// MarshalJSON is the default stringer method for the Details pointer.
func (d *Details) String() string {
	s, err := json.Marshal(d)
	if err != nil {
		return "error found."
	}
	return string(s)
}

// getGroupedValues converts a given slice into a GroupedValues object.
func getGroupedValues(data []float64) GroupedValues {
	var sum float64
	for i := range data {
		sum += data[i]
	}
	return GroupedValues{
		Sum:    roundOff(sum),
		Values: append([]float64{}, data...),
	}
}
