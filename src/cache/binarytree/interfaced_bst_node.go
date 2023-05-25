package binarytree

type IF_BSTNode[T Comparable[T]] struct {
	Val   T
	Left  *IF_BSTNode[T]
	Right *IF_BSTNode[T]
}

func (n *IF_BSTNode[T]) Value() T {
	return n.Val
}

func (n *IF_BSTNode[T]) insert(v T) (inserted bool) {
	if n.Val.Lt(v) {
		if n.Right == nil {
			n.Right = &IF_BSTNode[T]{Val: v}
			return true
		} else {
			return n.Right.insert(v)
		}
	} else if n.Val.Gt(v) {
		if n.Left == nil {
			n.Left = &IF_BSTNode[T]{Val: v}
			return true
		} else {
			return n.Left.insert(v)
		}
	} else if n.Val.Equals(v) {
		n.Val = v
	}
	return false
}

func (n *IF_BSTNode[T]) search(value T) (v T, ok bool) {
	// if we've reached the end of the tree, the value is not present
	if n == nil {
		return
	}
	// check if we need to traverse further down the tree
	if n.Val.Lt(value) {
		return n.Right.search(value)
	} else if n.Val.Gt(value) {
		return n.Left.search(value)
	} else if n.Val.Equals(value) {
		// value is not less than or greater than, so it must be equal
		return n.Val, true
	}
	return
}

func (n *IF_BSTNode[T]) traverse(f func(T)) {
	if n == nil {
		return
	}

	n.Left.traverse(f)
	f(n.Val)
	n.Right.traverse(f)
}

func (n *IF_BSTNode[T]) delete(v T) (newRoot *IF_BSTNode[T], deleted bool) {
	if n == nil {
		return nil, false
	}

	if v.Lt(n.Val) {
		n.Left, deleted = n.Left.delete(v)
	} else if v.Gt(n.Val) {
		n.Right, deleted = n.Right.delete(v)
	} else if v.Equals(n.Val) {
		deleted = true
		if n.Left == nil {
			return n.Right, deleted
		} else if n.Right == nil {
			return n.Left, deleted
		}

		minRight := n.Right.findMin()
		minRight.Right, _ = n.Right.delete(minRight.Val)
		minRight.Left = n.Left
		return minRight, deleted
	}

	return n, deleted
}

func (n *IF_BSTNode[T]) deleteIf(predicate func(T) bool) (newRoot *IF_BSTNode[T], deleted int) {
	if n == nil {
		return nil, 0
	}

	n.Left, deleted = n.Left.deleteIf(predicate)
	n.Right, deleted = n.Right.deleteIf(predicate)

	if predicate(n.Val) {
		deleted++
		if n.Left == nil {
			return n.Right, deleted
		} else if n.Right == nil {
			return n.Left, deleted
		}

		minRight := n.Right.findMin()
		minRight.Right, _ = n.Right.delete(minRight.Val)
		minRight.Left = n.Left
		return minRight, deleted
	}

	return n, deleted
}

func (n *IF_BSTNode[T]) findMin() *IF_BSTNode[T] {
	current := n
	for current.Left != nil {
		current = current.Left
	}
	return current
}

func (n *IF_BSTNode[T]) getHeight() int {
	if n == nil {
		return 0
	}

	leftHeight := n.Left.getHeight()
	rightHeight := n.Right.getHeight()

	if leftHeight > rightHeight {
		return leftHeight + 1
	}

	return rightHeight + 1
}
