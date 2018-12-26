// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.

package analytics

import (
	"errors"
)

// Insert appends every element in the source array into the binary tree with
// the first sourceArray element being made the root node. The sourceArray
// entries sum is used to determine which node to assign its data.
func (n *Node) Insert(sourceArray []GroupedValues) error {
	if n == nil {
		return errors.New("nil node cannot be assigned data")
	}

	n.Lock()
	defer n.Unlock()

	for i := range sourceArray {
		if i == 0 {
			// Assign the root node
			n.Value = sourceArray[i]
			continue
		}

		n.insert(sourceArray[i])
	}
	return nil
}

// insert is a recursive function that make the actual binary tree node add
// operation.
func (n *Node) insert(val GroupedValues) {
	switch {
	case val.Sum <= n.Value.Sum:
		if n.Left == nil {
			n.Left = &Node{Value: val}
			return
		}

		n.Left.insert(val)

	case val.Sum > n.Value.Sum:
		if n.Right == nil {
			n.Right = &Node{Value: val}
			return
		}
		n.Right.insert(val)
	}
}

// Transverse makes an inorder binary tree traversal returning a slice of all
// nodes data available.
func (n *Node) Transverse() []GroupedValues {
	var list []GroupedValues
	if n == nil {
		return list
	}

	output := make(chan GroupedValues)

	n.RLock()
	defer n.RUnlock()

	go func() {
		n.tranverse(output)
		close(output)
	}()

	for elem := range output {
		list = append(list, elem)
	}

	return list
}

// tranverse is the actual recursive function that walks through the provided
// binary tree sending out the nodes via a channel in Inorder binary tree traversal.
func (n *Node) tranverse(output chan<- GroupedValues) {
	if n.Left != nil {
		n.Left.tranverse(output)
	}

	output <- n.Value

	if n.Right != nil {
		n.Right.tranverse(output)
	}
}

// FindX returns all the matching values compared using the sum entry and an
// empty value if otherwise. Pre order binary tree traversal is used to
// avoid double matching.
func (n *Node) FindX(listX []GroupedValues, txFee float64) (matchingData []TxFundsFlow) {
	if n == nil {
		return
	}

	output := make(chan [2]GroupedValues)
	n.RLock()
	defer n.RUnlock()

	go func() {
		for i := range listX {
			n.findX(listX[i], output, txFee)
		}
		close(output)
	}()

	for elem := range output {
		matchingData = append(matchingData, TxFundsFlow{
			Fee:            roundOff(elem[0].Sum - elem[1].Sum),
			Inputs:         elem[0],
			MatchedOutputs: elem[1],
		})
	}

	return
}

// findX checks if a node entry whose comparison values match those in the
// provided input. If the matching node exists its data is returned.
func (n *Node) findX(val GroupedValues, output chan<- [2]GroupedValues, fee float64) {
	diff := roundOff(val.Sum - n.Value.Sum)
	if diff >= 0 && diff <= fee {
		output <- [2]GroupedValues{val, n.Value}
	}

	if n.Left != nil && diff < fee {
		n.Left.findX(val, output, fee)
	}

	if n.Right != nil && diff > 0 {
		n.Right.findX(val, output, fee)
	}
}
