// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package analytics

import (
	"math"
	"sort"
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

	for _, val := range list {
		s := d[val]
		if s == nil {
			s = &GroupedValues{}
		}
		sum = s.Sum
		sum += val
		s.Sum = roundOff(sum)
		s.Values = append(s.Values[:], val)
		d[val] = s
	}
	return d
}

// GenerateCombinations generates all the combinations for the array with the
// subset r count provided.
func GenerateCombinations(sourceArray []float64, r int64) []*GroupedValues {
	var newArrayIndex, oldArrayIndex, prevNewArrInd int64
	output := make(chan []float64)

	// Check if the slice is sorted.
	if !sort.Float64sAreSorted(sourceArray) {
		log.Warnf("Source array is not sorted. Total combinations maybe inaccurate.")
	}

	gD := GroupDuplicates(sourceArray)

	dups := make(map[float64]int, 0)
	// To increase the speed of iterative access to dups map contents
	// delete all entries without duplicates.
	for key, val := range gD {
		if len(val.Values) > 1 {
			dups[key] = len(val.Values)
		}
	}

	// Add the dope element when the last entry in the source slice is duplicate.
	lastEntry := sourceArray[len(sourceArray)-1]
	if _, ok := dups[lastEntry]; ok {
		sourceArray = append(sourceArray, dopingElement)
	}

	go func() {
		combinatorics(sourceArray, r, newArrayIndex, oldArrayIndex,
			prevNewArrInd, make([]float64, r, r), output, dups)
		close(output)
	}()

	res := make([]*GroupedValues, 0)

	for elem := range output {
		v := &GroupedValues{Values: elem}
		sum := 0.0
		for _, val := range elem {
			sum += val
			v.Sum = roundOff(sum)
		}

		res = append(res, v)
	}

	return res
}

// combinatorics is a recusive function that generates all the combinations C of
// subset r values from a set of n values. i.e nCr = n-1 C r-1 + n-1 C
func combinatorics(source []float64, r, newArrInd, sourceArrInd, prevNewArrInd int64,
	data []float64, output chan<- []float64, dups map[float64]int) {
	if newArrInd == r && data[len(data)-1] != dopingElement {
		tmp := make([]float64, len(data), len(data))
		copy(tmp, data)
		output <- tmp
		tmp = nil
		return
	}

	// value of sourceArrInd should always be less than the length of source
	// array to avoid 'index out of range' errors with source slice.
	if sourceArrInd >= int64(len(source)) {
		return
	}

	count := 1
	for i := 0; ; {
		// when r = 1 keep all the duplicates
		if r == 1 {
			break
		}

		if source[sourceArrInd] == data[newArrInd] {
			v := sourceArrInd + 1
			if v < int64(len(source)) {
				sourceArrInd = v
			}
		}

		count = dups[source[sourceArrInd]]
		if i >= count {
			break
		}
		i++
	}

	data[newArrInd] = source[sourceArrInd]
	prevNewArrInd = newArrInd

	combinatorics(source, r, newArrInd+1, sourceArrInd+1, prevNewArrInd, data, output, dups)
	combinatorics(source, r, newArrInd, sourceArrInd+1, prevNewArrInd, data, output, dups)
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
