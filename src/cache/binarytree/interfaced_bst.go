package binarytree

import (
	"fmt"
	"math"
	"strings"
)

type Equalizer[T any] interface {
	Equals(T) bool
}

type Comparable[T any] interface {
	Equalizer[T]
	Lt(T) bool
	Gt(T) bool
}

// A binary search tree implementation which works with any type that implements the Comparable[T] interface.
type InterfacedBST[T Comparable[T]] struct {
	Root   *IF_BSTNode[T]
	Length int
}

// Return the binary search tree as a string.
func (t *InterfacedBST[T]) String() string {
	if t.Root == nil {
		return ""
	}

	height := t.Root.getHeight()
	IF_BSTNodes := make([][]string, height)

	fillIF_BSTNodes(IF_BSTNodes, t.Root, 0)

	var b strings.Builder
	padding := int(math.Pow(2, float64(height)) - 1)

	for i, level := range IF_BSTNodes {
		if i == 0 {
			paddingStr := strings.Repeat(" ", (padding/2)+1)
			b.WriteString(paddingStr)
		} else {
			paddingStr := strings.Repeat(" ", padding/2)
			b.WriteString(paddingStr)
		}

		for j, IF_BSTNode := range level {
			b.WriteString(IF_BSTNode)
			if j != len(level)-1 {
				b.WriteString(strings.Repeat(" ", padding))
			}
		}

		padding /= 2
		b.WriteString("\n")
	}

	return b.String()
}

// Initialize a new binary search tree with the given initial value.
func NewInterfaced[T Comparable[T]](initial T) *InterfacedBST[T] {
	return &InterfacedBST[T]{
		Root: &IF_BSTNode[T]{Val: initial}}
}

// Insert a value into the binary search tree.
func (t *InterfacedBST[T]) Insert(value T) (inserted bool) {
	if t.Root == nil {
		t.Root = &IF_BSTNode[T]{Val: value}
		t.Length++
		return true
	}
	inserted = t.Root.insert(value)
	if inserted {
		t.Length++
	}
	return inserted
}

// Search for, and return, a value in the binary search tree.
func (t *InterfacedBST[T]) Search(value T) (v T, ok bool) {
	if t.Root == nil {
		return
	}
	return t.Root.search(value)
}

// Delete a value from the binary search tree.
func (t *InterfacedBST[T]) Delete(value T) (deleted bool) {
	if t.Root == nil {
		return false
	}
	t.Root, deleted = t.Root.delete(value)
	if deleted {
		t.Length--
	}
	return deleted
}

// Delete all values from the binary search tree that match the given predicate.
func (t *InterfacedBST[T]) DeleteIf(predicate func(T) bool) (deleted int) {
	if t.Root == nil {
		return 0
	}
	t.Root, deleted = t.Root.deleteIf(predicate)
	t.Length -= int(deleted)
	return deleted
}

// Traverse the binary search tree in-order.
func (t *InterfacedBST[T]) Traverse(f func(T)) {
	if t.Root == nil {
		return
	}
	t.Root.traverse(f)
}

// Return the number of values in the binary search tree.
func (t *InterfacedBST[T]) Len() int {
	return t.Length
}

// Return the height of the binary search tree.
func (t *InterfacedBST[T]) Height() int {
	if t.Root == nil {
		return 0
	}
	return t.Root.getHeight()
}

// Clear the binary search tree.
func (t *InterfacedBST[T]) Clear() {
	t.Root = nil
	t.Length = 0
}

func fillIF_BSTNodes[T Comparable[T]](IF_BSTNodes [][]string, n *IF_BSTNode[T], depth int) {
	if n == nil {
		return
	}

	IF_BSTNodes[depth] = append(IF_BSTNodes[depth], fmt.Sprintf("%v", n.Val))
	fillIF_BSTNodes(IF_BSTNodes, n.Left, depth+1)
	fillIF_BSTNodes(IF_BSTNodes, n.Right, depth+1)
}
