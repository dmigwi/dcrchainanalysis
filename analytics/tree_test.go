package analytics

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
)

// TestInsert runs tests on Insert method.
func TestInsert(t *testing.T) {
	type testData struct {
		StrJSONOutput string
		InputSlice    []GroupedValues
	}

	td := []testData{
		{
			StrJSONOutput: `{"Left":{"Value":{"Sum":11}},"Value":{"Sum":20}}`,
			InputSlice:    []GroupedValues{{Sum: 20}, {Sum: 11}},
		},
		{
			StrJSONOutput: `{"Left":{"Left":{"Value":{"Sum":1}},"Value":{"Sum":2}},` +
				`"Value":{"Sum":13},"Right":{"Value":{"Sum":17}}}`,
			InputSlice: []GroupedValues{{Sum: 13}, {Sum: 2}, {Sum: 17}, {Sum: 1}},
		},
		{
			StrJSONOutput: `{"Left":{"Value":{"Sum":1}},"Value":{"Sum":3},` +
				`"Right":{"Left":{"Value":{"Sum":4},"Right":{"Value":{"Sum":5}}},` +
				`"Value":{"Sum":6},"Right":{"Left":{"Value":{"Sum":7}},"Value":` +
				`{"Sum":7},"Right":{"Left":{"Left":{"Value":{"Sum":9}},"Value":` +
				`{"Sum":9}},"Value":{"Sum":10}}}}}`,
			InputSlice: []GroupedValues{
				{Sum: 3}, {Sum: 6}, {Sum: 7}, {Sum: 1}, {Sum: 4},
				{Sum: 10}, {Sum: 9}, {Sum: 5}, {Sum: 9}, {Sum: 7},
			},
		},
	}

	for i, data := range td {
		t.Run("Test_#"+strconv.Itoa(i+1), func(t *testing.T) {
			testTree := &Node{} // tree should be empty first

			if err := testTree.Insert(data.InputSlice); err != nil {
				t.Fatalf("expected no error to be returned but %v was returned", err)
			}

			// err return value is not significant since when its not nil, the
			// the test will definately fail.
			s, _ := json.Marshal(testTree)
			if string(s) != data.StrJSONOutput {
				t.Fatalf(`expected the returned JSON string to be (%s) but found (%s)`,
					data.StrJSONOutput, string(s))
			}
		})
	}
}

// TestTraverse tests the functionality of Traverse method.
func TestTraverse(t *testing.T) {
	type testData struct {
		InputSlice    []GroupedValues
		InorderOutput []GroupedValues
	}

	td := []testData{
		{
			InputSlice:    []GroupedValues{{Sum: 13}, {Sum: 2}, {Sum: 17}, {Sum: 1}},
			InorderOutput: []GroupedValues{{Sum: 1}, {Sum: 2}, {Sum: 13}, {Sum: 17}},
		},
		{
			InputSlice: []GroupedValues{{Sum: 10}, {Sum: 3}, {Sum: 6}, {Sum: 7},
				{Sum: 1}, {Sum: 4}, {Sum: 10}},
			InorderOutput: []GroupedValues{{Sum: 1}, {Sum: 3}, {Sum: 4}, {Sum: 6},
				{Sum: 7}, {Sum: 10}, {Sum: 10}},
		},
		{
			InputSlice: []GroupedValues{{Sum: 3}, {Sum: 6}, {Sum: 7}, {Sum: 1},
				{Sum: 4}, {Sum: 10}, {Sum: 9}, {Sum: 5}, {Sum: 9}, {Sum: 7}},
			InorderOutput: []GroupedValues{{Sum: 1}, {Sum: 3}, {Sum: 4}, {Sum: 5},
				{Sum: 6}, {Sum: 7}, {Sum: 7}, {Sum: 9}, {Sum: 9}, {Sum: 10}},
		},
	}

	for i, data := range td {
		t.Run("Test_#"+strconv.Itoa(i+1), func(t *testing.T) {
			testTree := &Node{}

			if err := testTree.Insert(data.InputSlice); err != nil {
				t.Fatalf("expected Insert not to return any err but found %v", err)
			}

			result := testTree.Transverse()
			if !reflect.DeepEqual(result, data.InorderOutput) {
				t.Fatalf("expected slice returned to be equal to (%v) but found (%v)",
					data.InorderOutput, result)
			}
		})
	}
}

// TestFindX tests the functionality of FindX.
func TestFindX(t *testing.T) {
	var testNilSlice []TxFundsFlow

	testTree := &Node{}
	inputSlice := []GroupedValues{{Sum: 13}, {Sum: 2}, {Sum: 17}, {Sum: 1}}

	type testData struct {
		ValueX    []GroupedValues
		MatchingX []TxFundsFlow
	}

	td := []testData{
		{[]GroupedValues{{Sum: 12}}, testNilSlice},
		{[]GroupedValues{{Sum: 13}}, []TxFundsFlow{
			{Fee: 0.0, Inputs: GroupedValues{Sum: 13}, MatchedOutputs: GroupedValues{Sum: 13}}},
		},
		{[]GroupedValues{{Sum: 17}}, []TxFundsFlow{
			{Fee: 0.0, Inputs: GroupedValues{Sum: 17}, MatchedOutputs: GroupedValues{Sum: 17}}},
		},
		{[]GroupedValues{{Sum: 19}}, testNilSlice},
	}

	if err := testTree.Insert(inputSlice); err != nil {
		t.Fatalf("expected Insert not to return any err but found %v", err)
	}

	for i, data := range td {
		t.Run("Test_#"+strconv.Itoa(i+1), func(t *testing.T) {
			result := testTree.FindX(data.ValueX, 0.0)
			if !reflect.DeepEqual(result, data.MatchingX) {
				t.Fatalf("expected X value to match %v but found %v", data.MatchingX, result)
			}
		})
	}
}
