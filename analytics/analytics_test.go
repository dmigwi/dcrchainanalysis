package analytics

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/raedahgroup/dcrchainanalysis/v1/rpcutils"
)

// TestTransactionFundsFlow tests the functionality of TransactionFundsFlow function.
func TestTransactionFundsFlow(t *testing.T) {
	txTestData := &rpcutils.Transaction{
		Fees:        0.000672,
		NumInpoint:  3,
		NumOutpoint: 4,
		Inpoints: []rpcutils.TxInput{
			{ValueIn: 39.96949337},
			{ValueIn: 40.9873785},
			{ValueIn: 5076.66042217},
		},
		Outpoints: []rpcutils.TxOutput{
			{Value: 39.96907437},
			{Value: 40.9873785},
			{Value: 40.9873785},
			{Value: 5035.67279067},
		},
	}

	expectedPayload := []*AllFundsFlows{
		{
			Solution:  1,
			TotalFees: 0.000672,
			FundsFlow: []TxFundsFlow{
				{
					Fee: 0.000419,
					Inputs: GroupedValues{
						Sum:    39.96949337,
						Values: []float64{39.96949337},
					},
					MatchedOutputs: GroupedValues{
						Sum:    39.96907437,
						Values: []float64{39.96907437},
					},
				},
				{
					Fee: 0,
					Inputs: GroupedValues{
						Sum:    40.9873785,
						Values: []float64{40.9873785},
					},
					MatchedOutputs: GroupedValues{
						Sum:    40.9873785,
						Values: []float64{40.9873785},
					},
				},
				{
					Fee: 0.000253,
					Inputs: GroupedValues{
						Sum: 5076.66042217, Values: []float64{5076.66042217},
					},
					MatchedOutputs: GroupedValues{
						Sum:    5076.66016917,
						Values: []float64{40.9873785, 5035.67279067},
					},
				},
			},
		},
	}

	t.Run("Test_#1", func(t *testing.T) {
		result, _, _, err := TransactionFundsFlow(txTestData)
		if err != nil {
			t.Fatalf("expected a nil value error to be returned but found: %v", err)
		}

		if len(result) != len(expectedPayload) {
			t.Fatalf("expected the returned results array to only have %v but it was %v",
				len(expectedPayload), len(result))
		}

		if isEquals := expectedPayload[0].equals(*result[0]); !isEquals {
			s, _ := json.Marshal(expectedPayload[0])
			q, _ := json.Marshal(*result[0])

			t.Fatalf("expected the returned payload to be equal"+
				" to (%v) but found (%v)", string(s), string(q))
		}
	})
}

// TestTxFundsFlowProbability tests the functionality of TxFundsFlowProbability function.
func TestTxFundsFlowProbability(t *testing.T) {
	txTestData := []*AllFundsFlows{
		{
			Solution:  1,
			TotalFees: 0.000672,
			FundsFlow: []TxFundsFlow{
				{
					Fee: 0.000419,
					Inputs: GroupedValues{
						Sum:    39.96949337,
						Values: []float64{39.96949337},
					},
					MatchedOutputs: GroupedValues{
						Sum:    39.96907437,
						Values: []float64{39.96907437},
					},
				},
				{
					Fee: 0,
					Inputs: GroupedValues{
						Sum:    40.9873785,
						Values: []float64{40.9873785},
					},
					MatchedOutputs: GroupedValues{
						Sum:    40.9873785,
						Values: []float64{40.9873785},
					},
				},
				{
					Fee: 0.000253,
					Inputs: GroupedValues{
						Sum:    5076.66042217,
						Values: []float64{5076.66042217},
					},
					MatchedOutputs: GroupedValues{
						Sum:    5076.66016917,
						Values: []float64{40.9873785, 5035.67279067},
					},
				},
			},
		},
	}

	expectedPayload := []*FlowProbability{
		{
			OutputAmount: 5035.67279067,
			Count:        1,
			ProbableInputs: []*InputSets{
				{Set: []*Details{{Amount: 5076.66042217, Count: 1}},
					PercentOfInputs: 1}},
			LinkingProbability: 1,
		},
		{
			OutputAmount: 39.96907437,
			Count:        1,
			ProbableInputs: []*InputSets{
				{Set: []*Details{{Amount: 39.96949337, Count: 1}},
					PercentOfInputs: 1}},
			LinkingProbability: 1,
		},
		{
			OutputAmount: 40.9873785,
			Count:        2,
			ProbableInputs: []*InputSets{
				{Set: []*Details{{Amount: 40.9873785, Count: 1}},
					PercentOfInputs: 1},
				{Set: []*Details{{Amount: 5076.66042217, Count: 1}},
					PercentOfInputs: 1}},
			LinkingProbability: 0.5,
		},
	}

	input := []float64{39.96949337, 40.9873785, 5076.66042217}
	output := []float64{39.96907437, 40.9873785, 40.9873785, 5035.67279067}

	t.Run("Test_#1", func(t *testing.T) {
		result := TxFundsFlowProbability(txTestData, input, output)

		if len(result) != len(expectedPayload) {
			t.Fatalf("expected the returned payload to be length %v but was %v",
				len(expectedPayload), len(result))
		}

	outerLoop:
		for _, expected := range expectedPayload {
			for _, returned := range result {
				if reflect.DeepEqual(expected.ProbableInputs, returned.ProbableInputs) {
					break outerLoop
				}

			}

			// If loop execution ever gets here then the test has already failed
			t.Fatal("expected the returned payload to be equal to" +
				" the returned payload but it wasn't")
		}
	})
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> <<<<<<<<<<<<<<<<<<<<<<<<<
// Benchmark tests

// BenchmarkTransactionFundsFlow is a benchmark test for TransactionFundsFlow.
func BenchmarkTransactionFundsFlow(b *testing.B) {
	txTestData := &rpcutils.Transaction{
		Fees:        0.000672,
		NumInpoint:  3,
		NumOutpoint: 4,
		Inpoints: []rpcutils.TxInput{
			{ValueIn: 39.96949337},
			{ValueIn: 40.9873785},
			{ValueIn: 5076.66042217},
		},
		Outpoints: []rpcutils.TxOutput{
			{Value: 39.96907437},
			{Value: 40.9873785},
			{Value: 40.9873785},
			{Value: 5035.67279067},
		},
	}

	for i := 0; i < b.N; i++ {
		TransactionFundsFlow(txTestData)
	}
}

// BenchmarkTxFundsFlowProbability is TxFundsFlowProbability benchmark test.
func BenchmarkTxFundsFlowProbability(b *testing.B) {
	txTestData := []*AllFundsFlows{
		{
			Solution:  1,
			TotalFees: 0.000672,
			FundsFlow: []TxFundsFlow{
				{
					Fee: 0.000419,
					Inputs: GroupedValues{
						Sum:    39.96949337,
						Values: []float64{39.96949337},
					},
					MatchedOutputs: GroupedValues{
						Sum:    39.96907437,
						Values: []float64{39.96907437},
					},
				},
				{
					Fee: 0,
					Inputs: GroupedValues{
						Sum:    40.9873785,
						Values: []float64{40.9873785},
					},
					MatchedOutputs: GroupedValues{
						Sum:    40.9873785,
						Values: []float64{40.9873785},
					},
				},
				{
					Fee: 0.000253,
					Inputs: GroupedValues{
						Sum:    5076.66042217,
						Values: []float64{5076.66042217},
					},
					MatchedOutputs: GroupedValues{
						Sum:    5076.66016917,
						Values: []float64{40.9873785, 5035.67279067},
					},
				},
			},
		},
	}

	input := []float64{39.96949337, 40.9873785, 5076.66042217}
	output := []float64{39.96907437, 40.9873785, 40.9873785, 5035.67279067}

	for i := 0; i < b.N; i++ {
		TxFundsFlowProbability(txTestData, input, output)
	}
}
