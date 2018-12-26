package analytics

import "sync"

const (
	// dopingElement is a placeholder value that helps guarrantee accuracy in
	// generating sum combinations with no duplicates when the source slice has
	// its last element as a duplicate.
	dopingElement float64 = -1

	// inpointData defines the input type of data.
	inpointData txProperties = "inputs"

	// outpointData defines the ouput type of data.
	outpointData txProperties = "outputs"
)

type txProperties string

// Node defines the basic unit of a binary tree. It has two children.
type Node struct {
	sync.RWMutex
	Left  *Node         `json:",omitempty"`
	Value GroupedValues `json:",omitempty"`
	Right *Node         `json:",omitempty"`
}

// GroupedValues clusters together values as duplicates or other grouped values.
// It holds the total sum and the list of the duplicates/grouped values.
type GroupedValues struct {
	Sum    float64   `json:",omitempty"`
	Values []float64 `json:",omitempty"`
}

// TxFundsFlow link inputs with their matching outputs. Also known as a Bucket.
type TxFundsFlow struct {
	Fee            float64
	Inputs         GroupedValues
	MatchedOutputs GroupedValues
}

// AllFundsFlows groups together all the possible solutions to the inputs and
// outputs funds flow.
type AllFundsFlows struct {
	Solution  int
	TotalFees float64
	FundsFlow []TxFundsFlow
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
	inputs          []float64
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
