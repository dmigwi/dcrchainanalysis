package analytics

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

// TestGeneratePermutations tests the functionality of the GeneratePermutations function.
func TestGeneratePermutations(t *testing.T) {
	type testData struct {
		N      int64
		R      int64
		Result int64
	}

	td := []testData{
		{N: 4, R: 2, Result: 12},
		{N: 4, R: 1, Result: 4},
		// Combinations with r < 1 always return 1 since its not allowed.
		{N: 4, R: 0, Result: 1},
		{N: 10, R: 3, Result: 720},
		{N: 15, R: 6, Result: 3603600},
		{N: 20, R: 10, Result: 670442572800},
	}

	for i, val := range td {
		t.Run("Test_#"+strconv.Itoa(i+1), func(t *testing.T) {
			result := GeneratePermutations(val.N, val.R)
			if result != val.Result {
				t.Fatalf("expected the C(%d, %d) to be equal to %d but %d was found",
					val.N, val.R, val.Result, result)
			}
		})
	}
}

// TestGroupDuplicates tests the functionality of
func TestGroupDuplicates(t *testing.T) {
	type testData struct {
		Input  []float64
		Output map[float64]*GroupedValues
	}

	td := []testData{
		{
			Input: []float64{1, 2, 3, 4},
			Output: map[float64]*GroupedValues{
				2: &GroupedValues{Sum: 2, Values: []float64{2}},
				3: &GroupedValues{Sum: 3, Values: []float64{3}},
				4: &GroupedValues{Sum: 4, Values: []float64{4}},
				1: &GroupedValues{Sum: 1, Values: []float64{1}},
			},
		},
		{
			Input: []float64{1, 3, 5, 7, 8, 1, 1, 1, 3, 3, 3, 4, 4, 4, 9, 9, 9, 9, 7, 5, 1, 3, 9},
			Output: map[float64]*GroupedValues{
				4: &GroupedValues{Sum: 12, Values: []float64{4, 4, 4}},
				9: &GroupedValues{Sum: 45, Values: []float64{9, 9, 9, 9, 9}},
				1: &GroupedValues{Sum: 5, Values: []float64{1, 1, 1, 1, 1}},
				3: &GroupedValues{Sum: 15, Values: []float64{3, 3, 3, 3, 3}},
				5: &GroupedValues{Sum: 10, Values: []float64{5, 5}},
				7: &GroupedValues{Sum: 14, Values: []float64{7, 7}},
				8: &GroupedValues{Sum: 8, Values: []float64{8}},
			},
		},
	}

	for i, data := range td {
		result := GroupDuplicates(data.Input)

		t.Run("Test_#"+strconv.Itoa(i+1), func(t *testing.T) {
			for key, val := range data.Output {

				re, ok := result[key]
				if !ok {
					t.Fatalf("expected key %v to be in %v but it wasn't", key, result)
				}

				if !reflect.DeepEqual(val, re) {
					t.Fatalf("expected slice (%v) to be equal to (%v) but it wasn't", re, val)
				}
			}
		})
	}
}

// TestAppendDupsCount tests the functionality of AppendDupsCount function.
func TestAppendDupsCount(t *testing.T) {
	type testData struct {
		TestArray []float64
		Results   []*Details
	}

	td := []*testData{
		&testData{
			TestArray: []float64{1, 2, 3, 4},
			Results: []*Details{
				{Amount: 1, Count: 1}, {Amount: 2, Count: 1}, {Amount: 3, Count: 1},
				{Amount: 4, Count: 1},
			},
		},
		&testData{
			TestArray: []float64{1, 1, 2, 2, 2, 3, 4, 4, 4},
			Results: []*Details{
				{Amount: 1, Count: 2}, {Amount: 1, Count: 2}, {Amount: 2, Count: 3},
				{Amount: 2, Count: 3}, {Amount: 2, Count: 3}, {Amount: 3, Count: 1},
				{Amount: 4, Count: 3}, {Amount: 4, Count: 3}, {Amount: 4, Count: 3},
				{Amount: -1, Count: 1},
			},
		},
		&testData{
			TestArray: []float64{1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 4, 4, 4, 5, 5, 7, 7, 8, 9, 9, 9, 9, 9},
			Results: []*Details{
				{Amount: 1, Count: 5}, {Amount: 1, Count: 5}, {Amount: 1, Count: 5},
				{Amount: 1, Count: 5}, {Amount: 1, Count: 5}, {Amount: 3, Count: 5},
				{Amount: 3, Count: 5}, {Amount: 3, Count: 5}, {Amount: 3, Count: 5},
				{Amount: 3, Count: 5}, {Amount: 4, Count: 3}, {Amount: 4, Count: 3},
				{Amount: 4, Count: 3}, {Amount: 5, Count: 2}, {Amount: 5, Count: 2},
				{Amount: 7, Count: 2}, {Amount: 7, Count: 2}, {Amount: 8, Count: 1},
				{Amount: 9, Count: 5}, {Amount: 9, Count: 5}, {Amount: 9, Count: 5},
				{Amount: 9, Count: 5}, {Amount: 9, Count: 5}, {Amount: -1, Count: 1},
			},
		},
	}

	for i, val := range td {
		t.Run("Test_#"+strconv.Itoa(i+1), func(t *testing.T) {
			res := appendDupsCount(val.TestArray)
			if !reflect.DeepEqual(res, val.Results) {
				expected, _ := json.Marshal(val.Results)
				returned, _ := json.Marshal(res)

				t.Fatalf("expected returned result to be equal to (%v) but found (%v)",
					string(expected), string(returned))
			}
		})
	}

}

// GenerateCombinations tests the functionality of GenerateCombinations function.
func TestGenerateCombinations(t *testing.T) {
	t.Parallel()

	type testData struct {
		TestArray []*Details
		R         int64
		Results   []*GroupedValues
	}

	td := []testData{
		{
			TestArray: []*Details{
				{Amount: 1, Count: 1}, {Amount: 2, Count: 1}, {Amount: 3, Count: 1},
				{Amount: 4, Count: 1},
			},
			R: 2,
			Results: []*GroupedValues{
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 5, Values: []float64{1, 4}},
				{Sum: 5, Values: []float64{2, 3}},
				{Sum: 6, Values: []float64{2, 4}},
				{Sum: 7, Values: []float64{3, 4}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 1, Count: 2}, {Amount: 1, Count: 2}, {Amount: 2, Count: 1},
				{Amount: 3, Count: 1},
			},
			R: 2,
			Results: []*GroupedValues{
				{Sum: 2, Values: []float64{1, 1}},
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 5, Values: []float64{2, 3}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 1, Count: 1}, {Amount: 2, Count: 2}, {Amount: 2, Count: 2},
				{Amount: 3, Count: 1},
			},
			R: 2,
			Results: []*GroupedValues{
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 4, Values: []float64{2, 2}},
				{Sum: 5, Values: []float64{2, 3}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 1, Count: 1}, {Amount: 2, Count: 3}, {Amount: 2, Count: 3},
				{Amount: 2, Count: 3}, {Amount: dopingElement, Count: 1},
			},
			R: 2,
			Results: []*GroupedValues{
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{2, 2}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 1, Count: 3}, {Amount: 1, Count: 3}, {Amount: 1, Count: 3},
				{Amount: 2, Count: 1}, {Amount: 3, Count: 1},
			},
			R: 2,
			Results: []*GroupedValues{
				{Sum: 2, Values: []float64{1, 1}},
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 5, Values: []float64{2, 3}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 1, Count: 3}, {Amount: 1, Count: 3}, {Amount: 1, Count: 3},
				{Amount: 2, Count: 3}, {Amount: 2, Count: 3}, {Amount: 2, Count: 3},
				{Amount: 3, Count: 1},
			},
			R: 2,
			Results: []*GroupedValues{
				{Sum: 2, Values: []float64{1, 1}},
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 4, Values: []float64{2, 2}},
				{Sum: 5, Values: []float64{2, 3}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 1, Count: 2}, {Amount: 1, Count: 2}, {Amount: 2, Count: 2},
				{Amount: 2, Count: 2}, {Amount: 3, Count: 2}, {Amount: 3, Count: 2},
				{Amount: dopingElement, Count: 1},
			},
			R: 2,
			Results: []*GroupedValues{
				{Sum: 2, Values: []float64{1, 1}},
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 4, Values: []float64{2, 2}},
				{Sum: 5, Values: []float64{2, 3}},
				{Sum: 6, Values: []float64{3, 3}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 1, Count: 2}, {Amount: 1, Count: 2}, {Amount: 2, Count: 3},
				{Amount: 2, Count: 3}, {Amount: 2, Count: 3}, {Amount: 3, Count: 1},
				{Amount: 4, Count: 3}, {Amount: 4, Count: 3}, {Amount: 4, Count: 3},
				{Amount: dopingElement, Count: 1},
			},
			R: 2,
			Results: []*GroupedValues{
				{Sum: 2, Values: []float64{1, 1}},
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 5, Values: []float64{1, 4}},
				{Sum: 4, Values: []float64{2, 2}},
				{Sum: 5, Values: []float64{2, 3}},
				{Sum: 6, Values: []float64{2, 4}},
				{Sum: 7, Values: []float64{3, 4}},
				{Sum: 8, Values: []float64{4, 4}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 5, Count: 2}, {Amount: 5, Count: 2}, {Amount: 6, Count: 2},
				{Amount: 6, Count: 2}, {Amount: 7, Count: 2}, {Amount: 7, Count: 2},
				{Amount: dopingElement, Count: 1},
			},
			R: 3,
			Results: []*GroupedValues{
				{Sum: 16, Values: []float64{5, 5, 6}},
				{Sum: 17, Values: []float64{5, 5, 7}},
				{Sum: 17, Values: []float64{5, 6, 6}},
				{Sum: 18, Values: []float64{5, 6, 7}},
				{Sum: 19, Values: []float64{5, 7, 7}},
				{Sum: 19, Values: []float64{6, 6, 7}},
				{Sum: 20, Values: []float64{6, 7, 7}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 5, Count: 2}, {Amount: 5, Count: 2}, {Amount: 6, Count: 2},
				{Amount: 6, Count: 2}, {Amount: 7, Count: 3}, {Amount: 7, Count: 3},
				{Amount: 7, Count: 3}, {Amount: dopingElement, Count: 1},
			},
			R: 3,
			Results: []*GroupedValues{
				{Sum: 16, Values: []float64{5, 5, 6}},
				{Sum: 17, Values: []float64{5, 5, 7}},
				{Sum: 17, Values: []float64{5, 6, 6}},
				{Sum: 18, Values: []float64{5, 6, 7}},
				{Sum: 19, Values: []float64{5, 7, 7}},
				{Sum: 19, Values: []float64{6, 6, 7}},
				{Sum: 20, Values: []float64{6, 7, 7}},
				{Sum: 21, Values: []float64{7, 7, 7}},
			},
		},
		{
			TestArray: []*Details{
				{Amount: 1, Count: 1}, {Amount: 2, Count: 1}, {Amount: 3, Count: 1},
				{Amount: 4, Count: 1}, {Amount: 5, Count: 1}, {Amount: 6, Count: 1},
			},
			R: 4,
			Results: []*GroupedValues{
				{Sum: 10, Values: []float64{1, 2, 3, 4}},
				{Sum: 11, Values: []float64{1, 2, 3, 5}},
				{Sum: 12, Values: []float64{1, 2, 3, 6}},
				{Sum: 12, Values: []float64{1, 2, 4, 5}},
				{Sum: 13, Values: []float64{1, 2, 4, 6}},
				{Sum: 14, Values: []float64{1, 2, 5, 6}},
				{Sum: 13, Values: []float64{1, 3, 4, 5}},
				{Sum: 14, Values: []float64{1, 3, 4, 6}},
				{Sum: 15, Values: []float64{1, 3, 5, 6}},
				{Sum: 16, Values: []float64{1, 4, 5, 6}},
				{Sum: 14, Values: []float64{2, 3, 4, 5}},
				{Sum: 15, Values: []float64{2, 3, 4, 6}},
				{Sum: 16, Values: []float64{2, 3, 5, 6}},
				{Sum: 17, Values: []float64{2, 4, 5, 6}},
				{Sum: 18, Values: []float64{3, 4, 5, 6}},
			},
		},
	}

	for i, val := range td {
		re := GenerateCombinations(val.TestArray, val.R)
		t.Run("Test_#"+strconv.Itoa(i+1), func(t *testing.T) {
			for i, d := range val.Results {
				c := re[i]
				if d.Sum != c.Sum {
					t.Fatalf("expected Sum to be %f but found %f at index %d", d.Sum, c.Sum, i)
				}

				if !reflect.DeepEqual(d.Values, c.Values) {
					t.Fatalf("expected grouped values to be %v but found %v at index %d",
						d.Values, c.Values, i)
				}
			}
		})
	}
}

// TestExtractSums tests the functionality of TestExtractSums
func TestExtractSums(t *testing.T) {
	type testData struct {
		Input      map[float64]*GroupedValues
		UniqueVals map[float64][]float64
		Output     []float64
	}

	td := []testData{
		{
			Input: map[float64]*GroupedValues{
				12.89:  &GroupedValues{Sum: 12.89, Values: []float64{10, 2.89}},
				332.89: &GroupedValues{Sum: 332.89, Values: []float64{300, 32.89}},
			},
			UniqueVals: map[float64][]float64{
				12.89:  []float64{10, 2.89},
				332.89: []float64{300, 32.89},
			},
			Output: []float64{12.89, 332.89},
		},
		{
			Input: map[float64]*GroupedValues{
				1: &GroupedValues{Sum: 2, Values: []float64{2}},
				2: &GroupedValues{Sum: 3, Values: []float64{1, 1, 1}},
				3: &GroupedValues{Sum: 4, Values: []float64{4}},
				5: &GroupedValues{Sum: 5, Values: []float64{2, 3}},
				6: &GroupedValues{Sum: 6, Values: []float64{3, 1, 2}},
			},
			UniqueVals: map[float64][]float64{
				2: []float64{2},
				3: []float64{1, 1, 1},
				4: []float64{4},
				5: []float64{2, 3},
				6: []float64{3, 1, 2},
			},
			Output: []float64{2, 3, 4, 5, 6},
		},
		{
			Input: map[float64]*GroupedValues{
				3478.89:  &GroupedValues{Sum: 36723478.8923, Values: []float64{2}},
				3283.90:  &GroupedValues{Sum: 328374928.90, Values: []float64{2}},
				5837.89:  &GroupedValues{Sum: 5892374.8, Values: []float64{2}},
				7847.896: &GroupedValues{Sum: 78347.896, Values: []float64{2}},
			},
			UniqueVals: map[float64][]float64{
				36723478.8923: []float64{2},
				328374928.90:  []float64{2},
				5892374.8:     []float64{2},
				78347.896:     []float64{2},
			},
			Output: []float64{78347.896, 5892374.8, 36723478.8923, 328374928.90},
		},
	}

	for i, data := range td {
		t.Run("Test_#"+strconv.Itoa(1+i), func(t *testing.T) {
			res, sumsToVals := ExtractSums(data.Input)

			// maps alter the original data entry order. Sort is needed to
			// ensure ascending order is maintained just for comparison purposes.
			// This sort function wasn't added to ExtractKeys because it will
			// introduce issues that will negatively affect performance with
			// no extra gain expected when using a sorted array instead of an
			// unsorted array on normal system functionality.
			sort.Float64s(res)

			if !reflect.DeepEqual(res, data.Output) {
				t.Fatalf("expected the returned slice to be equal to (%v) but found (%v)",
					data.Output, res)
			}

			if !reflect.DeepEqual(data.UniqueVals, sumsToVals) {
				t.Fatalf("expected the returned sums matched to unique values to be (%v) but found (%v)",
					data.UniqueVals, sumsToVals)
			}
		})
	}
}

// >>>>>>>>> <<<<<<<<<<<
// Benchmark tests

// benchmarkGenerateCombinations is a GenerateCombinations benchmark test
func benchmarkGenerateCombinations(arr []*Details, r int64, b *testing.B) {
	for n := 0; n < b.N; n++ {
		GenerateCombinations(arr, r)
	}
}

func BenchmarkGenerateCombinations1(b *testing.B) {
	arr := []*Details{
		{Amount: 1, Count: 1}, {Amount: 2, Count: 1}, {Amount: 3, Count: 1},
		{Amount: 4, Count: 1},
	}
	benchmarkGenerateCombinations(arr, 2, b)
}

func BenchmarkGenerateCombinations2(b *testing.B) {
	arr := []*Details{
		{Amount: 1, Count: 2}, {Amount: 1, Count: 2}, {Amount: 2, Count: 3},
		{Amount: 2, Count: 3}, {Amount: 2, Count: 3}, {Amount: 3, Count: 1},
		{Amount: 4, Count: 3}, {Amount: 4, Count: 3}, {Amount: 4, Count: 3},
		{Amount: dopingElement, Count: 1},
	}
	benchmarkGenerateCombinations(arr, 2, b)
}

func BenchmarkGenerateCombinations3(b *testing.B) {
	arr := []*Details{
		{Amount: 5, Count: 2}, {Amount: 5, Count: 2}, {Amount: 6, Count: 2},
		{Amount: 6, Count: 2}, {Amount: 7, Count: 3}, {Amount: 7, Count: 3},
		{Amount: 7, Count: 3}, {Amount: dopingElement, Count: 1},
	}
	benchmarkGenerateCombinations(arr, 3, b)
}

func BenchmarkGenerateCombinations4(b *testing.B) {
	arr := []*Details{
		{Amount: 1, Count: 1}, {Amount: 2, Count: 1}, {Amount: 3, Count: 1},
		{Amount: 4, Count: 1}, {Amount: 5, Count: 1}, {Amount: 6, Count: 1}}
	benchmarkGenerateCombinations(arr, 4, b)
}
func BenchmarkGenerateCombinations5(b *testing.B) {
	arr := []*Details{{Amount: 1, Count: 1}, {Amount: 2, Count: 1}, {Amount: 3, Count: 1}}
	benchmarkGenerateCombinations(arr, 2, b)
}
