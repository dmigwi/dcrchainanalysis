package analytics

import (
	"sync"
)

const (
	// dopingElement is a placeholder value that helps guarrantee accuracy in
	// generating sum combinations with no duplicates when the source slice has
	// its last element as a duplicate.
	dopingElement float64 = -1

	// inpointData defines the input type of data.
	inpointData txProperties = "inputs"

	// outpointData defines the ouput type of data.
	outpointData txProperties = "outputs"

	// txComplexityMeasure defines the number of unique transactions that may take
	// too long to process than needed. They may have a minimum probability of 5%.
	txComplexityMeasure = 20

	// complexTxMsg refers to the default message returned if a transaction cannot
	// be processed as fast as possible.
	complexTxMsg = "This txo is less then 5% traceable"
)

type txProperties string

// Node defines the basic unit of a binary tree. It has two children.
type Node struct {
	sync.RWMutex
	Left  *Node         `json:",omitempty"`
	Value GroupedValues `json:",omitempty"`
	Right *Node         `json:",omitempty"`
}

// Hub defines the basic unit of a transaction chain analysis graph. The Hub
// holds details of a funds flow linked between the current TxHash and the
// other Matched Hub(s). A chain of hubs provide the flow of funds from the current
// output TxHash to back in time where the source of funds can be identified.
type Hub struct {
	// Unique details of the current output.
	Address string
	Amount  float64
	TxHash  string

	// Probability Types
	PathProbability float64 `json:",omitempty"`
	hubProbability  float64

	// setCount helps track which set whose entry has already been processed
	// in a specific Hub.
	setCount int

	// Linked funds flow input(s).
	Matched []Set `json:",omitempty"`
}

// Set defines a group or individual inputs that can be correctly linked to an
// output as their source of funds.
type Set struct {
	// hubCount helps track which hub has already been processed.
	hubCount int

	Inputs          []*Hub
	PercentOfInputs float64
	StatusMsg       string `json:",omitempty"`
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
	TotalFees float64       `json:",omitempty"`
	FundsFlow []TxFundsFlow `json:",omitempty"`
	StatusMsg string        `json:",omitempty"`
}

// rawResults defines some compressed solutions data needed for further processing
// of the transaction funds flow.
type rawResults struct {
	Inputs          map[float64]int
	MatchingOutputs map[float64]*Details
}

// Details defines the input or output amount value and its duplicates count.
type Details struct {
	Amount         float64
	Count          int `json:",omitempty"`
	PossibleInputs int `json:",omitempty"`
	Actual         int `json:",omitempty"`
}

// InputSets groups probable inputs into sets each with its own percent of input value.
type InputSets struct {
	Set             []*Details
	PercentOfInputs float64
	inputs          []float64
	StatusMsg       string `json:",omitempty"`
}

// FlowProbability defines the final transaction funds flow data that includes
// the output tx funds flow probability.
type FlowProbability struct {
	OutputAmount       float64
	Count              int
	LinkingProbability float64
	ProbableInputs     []*InputSets
	uniqueInputs       map[float64]int
	StatusMsg          string `json:",omitempty"`
}

// custom sort interface that sorts by Possible inputs in the probability set
// data.
type byPossibleInputs []*Details

func (s byPossibleInputs) Len() int {
	return len(s)
}

func (s byPossibleInputs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byPossibleInputs) Less(i, j int) bool {
	return s[i].PossibleInputs > s[j].PossibleInputs
}
