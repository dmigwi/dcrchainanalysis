package analytics

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

// TestGetPrefabricatedBuckets tests the functionality of getPrefabricatedBuckets
// function.
func TestGetPrefabricatedBuckets(t *testing.T) {
	type testData struct {
		Array1    []float64
		Array2    []float64
		Buckets   []TxFundsFlow
		NewArray1 []float64
		NewArray2 []float64
	}

	td := []testData{
		{
			Array1: []float64{1, 1, 2, 4, 6, 7},
			Array2: []float64{1, 2, 3, 3, 4, 5, 6},
			Buckets: []TxFundsFlow{
				{
					Fee:            0,
					Inputs:         GroupedValues{Sum: 1, Values: []float64{1}},
					MatchedOutputs: GroupedValues{Sum: 1, Values: []float64{1}},
				},
				{
					Fee:            0,
					Inputs:         GroupedValues{Sum: 2, Values: []float64{2}},
					MatchedOutputs: GroupedValues{Sum: 2, Values: []float64{2}},
				},
				{
					Fee:            0,
					Inputs:         GroupedValues{Sum: 4, Values: []float64{4}},
					MatchedOutputs: GroupedValues{Sum: 4, Values: []float64{4}},
				},
				{
					Fee:            0,
					Inputs:         GroupedValues{Sum: 6, Values: []float64{6}},
					MatchedOutputs: GroupedValues{Sum: 6, Values: []float64{6}},
				},
			},
			NewArray1: []float64{1, 7},
			NewArray2: []float64{3, 3, 5},
		},
	}

	for i, data := range td {
		t.Run("Test_"+strconv.Itoa(i), func(t *testing.T) {
			b, arr1, arr2 := getPrefabricatedBuckets(data.Array1, data.Array2)

			if !reflect.DeepEqual(data.NewArray1, arr1) {
				t.Fatalf("expected NewArray1 returned to be equal to (%v) but found (%v)",
					data.NewArray1, arr1)
			}

			if !reflect.DeepEqual(data.NewArray2, arr2) {
				t.Fatalf("expected NewArray2 returned to be equal to (%v) but found (%v)",
					data.NewArray2, arr2)
			}

			if !reflect.DeepEqual(data.Buckets, b) {
				s, _ := json.Marshal(data.Buckets)
				p, _ := json.Marshal(b)
				t.Fatalf("expected Buckets returned to be equal to (%v) but found (%v)",
					string(s), string(p))
			}
		})
	}
}
