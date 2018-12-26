package analytics

import (
	"encoding/json"
	"reflect"
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
		Output map[float64]GroupedValues
	}

	td := []testData{
		{
			Input: []float64{1, 2, 3, 4},
			Output: map[float64]GroupedValues{
				2: GroupedValues{Sum: 2, Values: []float64{2}},
				3: GroupedValues{Sum: 3, Values: []float64{3}},
				4: GroupedValues{Sum: 4, Values: []float64{4}},
				1: GroupedValues{Sum: 1, Values: []float64{1}},
			},
		},
		{
			Input: []float64{1, 3, 5, 7, 8, 1, 1, 1, 3, 3, 3, 4, 4, 4, 9, 9, 9, 9, 7, 5, 1, 3, 9},
			Output: map[float64]GroupedValues{
				4: GroupedValues{Sum: 12, Values: []float64{4, 4, 4}},
				9: GroupedValues{Sum: 45, Values: []float64{9, 9, 9, 9, 9}},
				1: GroupedValues{Sum: 5, Values: []float64{1, 1, 1, 1, 1}},
				3: GroupedValues{Sum: 15, Values: []float64{3, 3, 3, 3, 3}},
				5: GroupedValues{Sum: 10, Values: []float64{5, 5}},
				7: GroupedValues{Sum: 14, Values: []float64{7, 7}},
				8: GroupedValues{Sum: 8, Values: []float64{8}},
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
				{Amount: 9, Count: 5}, {Amount: 9, Count: 5},
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
		TestArray []float64
		R         int64
		Results   []GroupedValues
	}

	td := []testData{
		{
			TestArray: []float64{1, 2, 3, 4},
			R:         2,
			Results: []GroupedValues{
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 5, Values: []float64{1, 4}},
				{Sum: 5, Values: []float64{2, 3}},
				{Sum: 6, Values: []float64{2, 4}},
				{Sum: 7, Values: []float64{3, 4}},
			},
		},
		{
			TestArray: []float64{1, 1, 2, 3},
			R:         2,
			Results: []GroupedValues{
				{Sum: 2, Values: []float64{1, 1}},
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 5, Values: []float64{2, 3}},
			},
		},
		{
			TestArray: []float64{1, 2, 2, 3},
			R:         2,
			Results: []GroupedValues{
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 4, Values: []float64{2, 2}},
				{Sum: 5, Values: []float64{2, 3}},
			},
		},
		{
			TestArray: []float64{1, 2, 2, 2, dopingElement},
			R:         2,
			Results: []GroupedValues{
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{2, 2}},
			},
		},
		{
			TestArray: []float64{1, 1, 1, 2, 3},
			R:         2,
			Results: []GroupedValues{
				{Sum: 2, Values: []float64{1, 1}},
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 5, Values: []float64{2, 3}},
			},
		},
		{
			TestArray: []float64{1, 1, 1, 2, 2, 2, 3},
			R:         2,
			Results: []GroupedValues{
				{Sum: 2, Values: []float64{1, 1}},
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 4, Values: []float64{2, 2}},
				{Sum: 5, Values: []float64{2, 3}},
			},
		},
		{
			TestArray: []float64{1, 1, 2, 2, 3, 3, dopingElement},
			R:         2,
			Results: []GroupedValues{
				{Sum: 2, Values: []float64{1, 1}},
				{Sum: 3, Values: []float64{1, 2}},
				{Sum: 4, Values: []float64{1, 3}},
				{Sum: 4, Values: []float64{2, 2}},
				{Sum: 5, Values: []float64{2, 3}},
				{Sum: 6, Values: []float64{3, 3}},
			},
		},
		{
			TestArray: []float64{1, 1, 2, 2, 2, 3, 4, 4, 4, dopingElement},
			R:         2,
			Results: []GroupedValues{
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
			TestArray: []float64{5, 5, 6, 6, 7, 7, dopingElement},
			R:         3,
			Results: []GroupedValues{
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
			TestArray: []float64{5, 5, 6, 6, 7, 7, 7, dopingElement},
			R:         3,
			Results: []GroupedValues{
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
			TestArray: []float64{1, 2, 3, 4, 5, 6},
			R:         4,
			Results: []GroupedValues{
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

// TestIsEqual tests the functionality of isEqual function.
func TestIsEqual(t *testing.T) {
	type testData struct {
		Array1 []float64
		Array2 []float64
	}

	td := []testData{
		{
			Array1: []float64{4, 5, 6, 7, 8, 9, 10, 11, 12},
			Array2: []float64{4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			Array1: []float64{5, 6, 7, 8, 9, 10, 11, 12, 4},
			Array2: []float64{4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			Array1: []float64{6, 7, 8, 9, 10, 11, 12, 4, 5},
			Array2: []float64{7, 8, 9, 10, 11, 12, 4, 5},
		},
		{
			Array1: []float64{7, 8, 9, 10, 11, 12, 4, 5, 6},
			Array2: []float64{7, 8, 9, 10, 11, 12, 4, 5, 6},
		},
		{
			Array1: []float64{8, 9, 10, 11, 12, 4, 5, 6, 7},
			Array2: []float64{8, 9, 10, 11, 12, 4, 5, 6, 7},
		},
		{
			Array1: []float64{9, 10, 11, 12, 4, 5, 6, 7, 8},
			Array2: []float64{9, 10, 11, 12, 4, 5, 6, 7, 8},
		},
		{
			Array1: []float64{10, 11, 12, 4, 5, 6, 7, 8, 9},
			Array2: []float64{10, 11, 12, 5, 4, 6, 7, 8, 9},
		},
		{
			Array1: []float64{11, 12, 4, 5, 6, 7, 8, 9, 10},
			Array2: []float64{11, 12, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			Array1: []float64{12, 4, 5, 6, 7, 8, 9, 10, 11},
			Array2: []float64{12, 4, 5, 6, 7, 8, 9, 10, 11},
		},
	}

	for i, data := range td {
		t.Run("Test_#"+strconv.Itoa(i+1), func(t *testing.T) {
			res := isEqual(data.Array1, data.Array2)
			accurateResult := reflect.DeepEqual(data.Array1, data.Array2)

			if res != accurateResult {
				t.Fatalf("slice equality check for (%v) and (%v) failed",
					data.Array1, data.Array2)
			}
		})
	}
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
// Benchmark tests

// benchmarkGenerateCombinations is a GenerateCombinations benchmark test
func benchmarkGenerateCombinations(arr []float64, r int64, b *testing.B) {
	for n := 0; n < b.N; n++ {
		GenerateCombinations(arr, r)
	}
}

func BenchmarkGenerateCombinations1(b *testing.B) {
	arr := []float64{1, 2, 3, 4}
	benchmarkGenerateCombinations(arr, 2, b)
}

func BenchmarkGenerateCombinations2(b *testing.B) {
	arr := []float64{1, 1, 2, 2, 2, 3, 4, 4, 4, dopingElement}
	benchmarkGenerateCombinations(arr, 2, b)
}

func BenchmarkGenerateCombinations3(b *testing.B) {
	arr := []float64{5, 5, 6, 6, 7, 7, 7, dopingElement}
	benchmarkGenerateCombinations(arr, 3, b)
}

func BenchmarkGenerateCombinations4(b *testing.B) {
	arr := []float64{1, 2, 3, 4, 5, 6}
	benchmarkGenerateCombinations(arr, 4, b)
}
func BenchmarkGenerateCombinations5(b *testing.B) {
	arr := []float64{1, 2, 3}
	benchmarkGenerateCombinations(arr, 2, b)
}
