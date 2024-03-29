// avl.go - An AVL tree implementation.
//
// To the extent possible under law, Jesse Kuang has waived all copyright
// and related or neighboring rights to go-avl, using the Creative
// Commons "CC0" public domain dedication. See LICENSE for full details.
//
// To the extent possible under law, Yawning Angel has waived all copyright
// and related or neighboring rights to avl, using the Creative
// Commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

// Package avl implements an AVL tree.
package avl

// This is a fairly straight forward adaptation of the CC0 C implementation
// from https://github.com/ebiggers/avl_tree/ by Eric Biggers into what is
// hopefully idiomatic Go.
//
// The primary differences from the original package are:
//  * The balance factor is not stored separately from the parent pointer.
//  * The container is non-intrusive.
//  * Only in-order traversal is currently supported.

import "errors"

var (
	errNoCmpFn          = errors.New("avl: no comparison function")
	errNotInTree        = errors.New("avl: element not in tree")
	errInvalidDirection = errors.New("avl: invalid direction")
)

// Comparable is Interface the Cmp function used to compare entries in the
// Tree to maintain ordering.  It MUST return < 0, 0, or > 0 if the a is less
// than, equal to, or greater than b respectively.
//
// Note: All calls made to the comparison function will pass the user supplied
// value as a, and the in-Tree value as b.
type CompareFunc[K any] func(a, b K) int

// Direction is the direction associated with an iterator.
type Direction int

const (
	// Forward is forward in-order.
	Forward Direction = 1
	it_sign int       = 1
)

// Node is a node of a Tree.
type Node[T, K any] struct {
	// Value is the value stored by the Node.
	Value T
	Key   K

	parent, left, right *Node[T, K]
	balance             int
}

// Iterator is a Tree iterator.  Modifying the Tree while iterating is
// unsupported except for removing the current Node.
type Iterator[T, K any] struct {
	tree      *Tree[T, K]
	cur, next *Node[T, K]
	//sign        int
	initialized bool
}

// First moves the iterator to the first Node in the Tree and returns the
// first Node or nil iff the Tree is empty.  Note that "first" in this context
// is dependent on the direction specified when constructing the iterator.
func (it *Iterator[T, K]) First() *Node[T, K] {
	it.cur, it.next = it.tree.First(), nil
	if it.cur != nil {
		it.next = it.cur.nextOrPrevInOrder(it_sign)
	}
	it.initialized = true
	return it.cur
}

// Get returns the Node currently pointed to by the iterator.  It is safe to
// remove the Node returned from the Tree.
func (it *Iterator[T, K]) Get() *Node[T, K] {
	if !it.initialized {
		return it.First()
	}
	return it.cur
}

// Next advances the iterator and returns the Node or nil iff the end of the
// Tree has been reached.
func (it *Iterator[T, K]) Next() *Node[T, K] {
	if !it.initialized {
		it.First()
	}
	it.cur = it.next
	if it.next == nil {
		return nil
	}

	it.next = it.cur.nextOrPrevInOrder(it_sign)
	return it.cur
}

func (n *Node[T, K]) reset() {
	// Note: This deliberately leaves Value intact.
	n.parent, n.left, n.right = n, nil, nil
	n.balance = 0
}

func (n *Node[T, K]) setParentBalance(parent *Node[T, K], balance int) {
	n.parent = parent
	n.balance = balance
}

func (n *Node[T, K]) getChild(sign int) *Node[T, K] {
	if sign < 0 {
		return n.left
	}
	return n.right
}

func (n *Node[T, K]) nextOrPrevInOrder(sign int) *Node[T, K] {
	var next, tmp *Node[T, K]

	if next = n.getChild(+sign); next != nil {
		for {
			tmp = next.getChild(-sign)
			if tmp == nil {
				break
			}
			next = tmp
		}
	} else {
		tmp, next = n, n.parent
		for next != nil && tmp == next.getChild(+sign) {
			tmp, next = next, next.parent
		}
	}

	return next
}

func (n *Node[T, K]) setChild(sign int, child *Node[T, K]) {
	if sign < 0 {
		n.left = child
	} else {
		n.right = child
	}
}

func (n *Node[T, K]) adjustBalanceFactor(amount int) {
	n.balance += amount
}

// Tree represents an AVL tree.
type Tree[T, K any] struct {
	root  *Node[T, K]
	first *Node[T, K]
	cmpFn CompareFunc[K]
	size  int
}

// Len returns the number of elements in the Tree.
func (t *Tree[T, K]) Len() int {
	return t.size
}

// First returns the first node in the Tree (in-order) or nil iff the Tree is
// empty.
func (t *Tree[T, K]) First() *Node[T, K] {
	if t.first == nil {
		t.first = t.firstOrLastInOrder(-1)
	}
	return t.first

}

// Last returns the last element in the Tree (in-order) or nil iff the Tree is
// empty.
func (t *Tree[T, K]) Last() *Node[T, K] {
	return t.firstOrLastInOrder(1)
}

// Find finds the value in the Tree, and returns the Node or nil iff the value
// is not present.
func (t *Tree[T, K]) Find(key K) *Node[T, K] {
	if t.cmpFn == nil {
		panic(errNoCmpFn)
	}

	cur := t.root
descendLoop:
	for cur != nil {
		cmp := t.cmpFn(key, cur.Key)
		switch {
		case cmp < 0:
			cur = cur.left
		case cmp > 0:
			cur = cur.right
		default:
			break descendLoop
		}
	}

	return cur
}

// Insert inserts the value into the Tree, and returns the newly created Node
// or the existing Node iff the value is already present in the tree.
func (t *Tree[T, K]) Insert(k K, v T) *Node[T, K] {
	if t.cmpFn == nil {
		panic(errNoCmpFn)
	}

	var cur *Node[T, K]
	if t.first != nil {
		if t.cmpFn(k, t.first.Key) < 0 {
			t.first = nil
		}
	}
	curPtr := &t.root
	for *curPtr != nil {
		cur = *curPtr
		cmp := t.cmpFn(k, cur.Key)
		switch {
		case cmp < 0:
			curPtr = &cur.left
		case cmp > 0:
			curPtr = &cur.right
		default:
			return cur
		}
	}

	n := &Node[T, K]{
		Value:   v,
		Key:     k,
		parent:  cur,
		balance: 0,
	}
	*curPtr = n
	t.rebalanceAfterInsert(n)
	t.size++

	return n
}

// Insert inserts the Node into the Tree, and returns the Node
// or null if the node value is already present in the tree.
func (t *Tree[T, K]) InsertNode(n *Node[T, K]) *Node[T, K] {
	k := n.Key
	if t.cmpFn == nil {
		panic(errNoCmpFn)
	}

	var cur *Node[T, K]
	if t.first != nil {
		if t.cmpFn(k, t.first.Key) < 0 {
			t.first = nil
		}
	}
	curPtr := &t.root
	for *curPtr != nil {
		cur = *curPtr
		cmp := t.cmpFn(k, cur.Key)
		switch {
		case cmp < 0:
			curPtr = &cur.left
		case cmp > 0:
			curPtr = &cur.right
		default:
			return cur
		}
	}

	// initialize node pointers
	n.reset()
	n.parent = cur
	*curPtr = n
	t.rebalanceAfterInsert(n)
	t.size++

	return n
}

// Remove removes the Node from the Tree.
func (t *Tree[T, K]) Remove(node *Node[T, K]) {
	var parent *Node[T, K]
	var leftDeleted bool

	if node.parent == node {
		panic(errNotInTree)
	}
	if t.first != nil {
		if t.cmpFn(node.Key, t.first.Key) <= 0 {
			t.first = nil
		}
	}

	t.size--
	if node.left != nil && node.right != nil {
		parent, leftDeleted = t.swapWithSuccessor(node)
	} else {
		child := node.left
		if child == nil {
			child = node.right
		}
		parent = node.parent
		if parent != nil {
			if node == parent.left {
				parent.left = child
				leftDeleted = true
			} else {
				parent.right = child
				leftDeleted = false
			}
			if child != nil {
				child.parent = parent
			}
		} else {
			if child != nil {
				child.parent = parent
			}
			t.root = child
			node.reset()
			return
		}
	}

	for {
		if leftDeleted {
			parent = t.handleSubtreeShrink(parent, +1, &leftDeleted)
		} else {
			parent = t.handleSubtreeShrink(parent, -1, &leftDeleted)
		}
		if parent == nil {
			break
		}
	}
	node.reset()
}

// Iterator returns an iterator that traverses the tree (in-order) in the
// specified direction.  Modifying the Tree while iterating is unsupported
// except for removing the current Node.
func (t *Tree[T, K]) Iterator(direction Direction) *Iterator[T, K] {
	switch direction {
	case Forward:
	default:
		panic(errInvalidDirection)
	}

	return &Iterator[T, K]{
		tree: t,
		//sign: 1, //int(direction),
	}
}

// ForEach executes a function for each Node in the tree, visiting the nodes
// in-order in the direction specified.  If the provided function returns
// false, the iteration is stopped.  Modifying the Tree from within the
// function is unsupprted except for removing the current Node.
func (t *Tree[T, K]) ForEach(direction Direction, fn func(*Node[T, K]) bool) {
	it := t.Iterator(direction)
	for node := it.Get(); node != nil; node = it.Next() {
		if !fn(node) {
			return
		}
	}
}

func (t *Tree[T, K]) firstOrLastInOrder(sign int) *Node[T, K] {
	first := t.root
	if first != nil {
		for {
			tmp := first.getChild(+sign)
			if tmp == nil {
				break
			}
			first = tmp
		}
	}
	return first
}

func (t *Tree[T, K]) replaceChild(parent, oldChild, newChild *Node[T, K]) {
	if parent != nil {
		if oldChild == parent.left {
			parent.left = newChild
		} else {
			parent.right = newChild
		}
	} else {
		t.root = newChild
	}
}

func (t *Tree[T, K]) rotate(a *Node[T, K], sign int) {
	b := a.getChild(-sign)
	e := b.getChild(+sign)
	p := a.parent

	a.setChild(-sign, e)
	a.parent = b

	b.setChild(+sign, a)
	b.parent = p

	if e != nil {
		e.parent = a
	}

	t.replaceChild(p, a, b)
}

func (t *Tree[T, K]) doDoubleRotate(b, a *Node[T, K], sign int) *Node[T, K] {
	e := b.getChild(+sign)
	f := e.getChild(-sign)
	g := e.getChild(+sign)
	p := a.parent
	eBal := e.balance

	a.setChild(-sign, g)
	aBal := -eBal
	if sign*eBal >= 0 {
		aBal = 0
	}
	a.setParentBalance(e, aBal)

	b.setChild(+sign, f)
	bBal := -eBal
	if sign*eBal <= 0 {
		bBal = 0
	}
	b.setParentBalance(e, bBal)

	e.setChild(+sign, a)
	e.setChild(-sign, b)
	e.setParentBalance(p, 0)

	if g != nil {
		g.parent = a
	}

	if f != nil {
		f.parent = b
	}

	t.replaceChild(p, a, e)

	return e
}

func (t *Tree[T, K]) handleSubtreeGrowth(node, parent *Node[T, K], sign int) bool {
	oldBalanceFactor := parent.balance
	if oldBalanceFactor == 0 {
		parent.adjustBalanceFactor(sign)
		return false
	}

	newBalanceFactor := oldBalanceFactor + sign
	if newBalanceFactor == 0 {
		parent.adjustBalanceFactor(sign)
		return true
	}

	if sign*node.balance > 0 {
		t.rotate(parent, -sign)
		parent.adjustBalanceFactor(-sign)
		node.adjustBalanceFactor(-sign)
	} else {
		t.doDoubleRotate(node, parent, -sign)
	}

	return true
}

func (t *Tree[T, K]) rebalanceAfterInsert(inserted *Node[T, K]) {
	node, parent := inserted, inserted.parent
	switch {
	case parent == nil:
		return
	case node == parent.left:
		parent.adjustBalanceFactor(-1)
	default:
		parent.adjustBalanceFactor(+1)
	}

	if parent.balance == 0 {
		return
	}

	for done := false; !done; {
		node = parent
		if parent = node.parent; parent == nil {
			return
		}

		if node == parent.left {
			done = t.handleSubtreeGrowth(node, parent, -1)
		} else {
			done = t.handleSubtreeGrowth(node, parent, +1)
		}
	}
}

func (t *Tree[T, K]) swapWithSuccessor(x *Node[T, K]) (*Node[T, K], bool) {
	var ret *Node[T, K]
	var leftDeleted bool

	y := x.right
	if y.left == nil {
		ret = y
	} else {
		var q *Node[T, K]

		for {
			q = y
			if y = y.left; y.left == nil {
				break
			}
		}

		if q.left = y.right; q.left != nil {
			q.left.parent = q
		}
		y.right = x.right
		x.right.parent = y
		ret = q
		leftDeleted = true
	}

	y.left = x.left
	x.left.parent = y

	y.parent = x.parent
	y.balance = x.balance

	t.replaceChild(x.parent, x, y)

	return ret, leftDeleted
}

func (t *Tree[T, K]) handleSubtreeShrink(parent *Node[T, K], sign int,
	leftDeleted *bool) *Node[T, K] {
	oldBalanceFactor := parent.balance
	if oldBalanceFactor == 0 {
		parent.adjustBalanceFactor(sign)
		return nil
	}

	var node *Node[T, K]
	newBalanceFactor := oldBalanceFactor + sign
	if newBalanceFactor == 0 {
		parent.adjustBalanceFactor(sign)
		node = parent
	} else {
		node = parent.getChild(sign)
		if sign*node.balance >= 0 {
			t.rotate(parent, -sign)
			if node.balance == 0 {
				node.adjustBalanceFactor(-sign)
				return nil
			}
			parent.adjustBalanceFactor(-sign)
			node.adjustBalanceFactor(-sign)
		} else {
			node = t.doDoubleRotate(node, parent, -sign)
		}
	}
	if parent = node.parent; parent != nil {
		*leftDeleted = node == parent.left
	}
	return parent
}

// New returns an initialized Tree.
func New[T, K any](cmpFn CompareFunc[K]) *Tree[T, K] {
	if cmpFn == nil {
		panic(errNoCmpFn)
	}

	return &Tree[T, K]{cmpFn: cmpFn}
}
