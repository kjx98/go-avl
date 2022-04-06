// avl_test.go - AVL tree tests.
//
// To the extent possible under law, Jesse Kuang has waived all copyright
// and related or neighboring rights to go-avl, using the Creative
// Commons "CC0" public domain dedication. See LICENSE for full details.
//
// To the extent possible under law, Yawning Angel has waived all copyright
// and related or neighboring rights to avl, using the Creative
// Commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package avl

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestAVLTree(t *testing.T) {
	Equal := func(a, b any, ss string, args ...any) {
		if !reflect.DeepEqual(a, b) {
			t.Errorf(ss, args...)
		}
	}
	Nil := func(a any, ss string, args ...any) {
		if !reflect.ValueOf(a).IsNil() {
			t.Errorf(ss, args...)
		}
	}

	tree := New[int](func(a, b int) int {
		return a - b
	})
	Equal(0, tree.Len(), "Len(): empty")
	Nil(tree.First(), "First(): empty")
	Nil(tree.Last(), "Last(): empty")

	iter := tree.Iterator(Forward)
	Nil(iter.First(), "Iterator: First(), empty")
	Nil(iter.Next(), "Iterator: Next(), empty")

	// Test insertion.
	const nrEntries = 1024
	insertedMap := make(map[int]*Node[int])
	for len(insertedMap) != nrEntries {
		v := rand.Int()
		if insertedMap[v] != nil {
			continue
		}
		insertedMap[v] = tree.Insert(int(v))
		tree.validate(t)
	}
	Equal(nrEntries, tree.Len(), "Len(): After insertion")
	t.Log("After insertion OK")
	tree.validate(t)
	t.Log("After validate OK")

	// Ensure that all entries can be found.
	for k, v := range insertedMap {
		Equal(v, tree.Find(k), "Find(): %v", k)
		Equal(k, int(v.Value), "Find(): %v Value", k)
	}

	// Test the forward/backward iterators.
	fwdInOrder := make([]int, 0, nrEntries)
	for k := range insertedMap {
		fwdInOrder = append(fwdInOrder, int(k))
	}
	sort.Ints(fwdInOrder)
	Equal(fwdInOrder[0], tree.First().Value, "First(), full")
	Equal(fwdInOrder[nrEntries-1], tree.Last().Value, "Last(), full")
	t.Log("After forward validate OK")

	revInOrder := make([]int, 0, nrEntries)
	for i := len(fwdInOrder) - 1; i >= 0; i-- {
		revInOrder = append(revInOrder, fwdInOrder[i])
	}

	iter = tree.Iterator(Forward)
	visited := 0
	for node := iter.First(); node != nil; node = iter.Next() {
		v, idx := node.Value, visited
		Equal(fwdInOrder[visited], int(v), "Iterator: Forward[%v]", idx)
		Equal(node, iter.Get(), "Iterator: Forward[%v]: Get()", idx)
		visited++
	}
	Equal(nrEntries, visited, "Iterator: Forward: Visited")
	t.Log("After Iterator: Forward: Visited")

	// Test the forward/backward ForEach.
	forEachValues := make([]int, 0, nrEntries)
	forEachFn := func(n *Node[int]) bool {
		forEachValues = append(forEachValues, int(n.Value))
		return true
	}
	tree.ForEach(Forward, forEachFn)
	Equal(fwdInOrder, forEachValues, "ForEach: Forward")
	t.Log("After ForEach: Forward")

	forEachValues = make([]int, 0, nrEntries)

	// Test removal.
	for i, idx := range rand.Perm(nrEntries) { // In random order.
		v := fwdInOrder[idx]
		node := tree.Find(int(v))
		Equal(v, int(node.Value), "Find(): %v (Pre-remove)", v)

		tree.Remove(node)
		Equal(nrEntries-(i+1), tree.Len(), "Len(): %v (Post-remove)", v)
		tree.validate(t)

		node = tree.Find(int(v))
		Nil(node, "Find(): %v (Post-remove)", v)
	}
	Equal(0, tree.Len(), "Len(): After removal")
	Nil(tree.First(), "First(): After removal")
	Nil(tree.Last(), "Last(): After removal")

	// Refill the tree.
	for _, v := range fwdInOrder {
		tree.Insert(int(v))
	}

	// Test that removing the node doesn't break the iterator.
	iter = tree.Iterator(Forward)
	visited = 0
	for node := iter.Get(); node != nil; node = iter.Next() { // Omit calling First().
		v, idx := node.Value, visited
		Equal(fwdInOrder[idx], int(v), "Iterator: Forward[%v] (Pre-Remove)", idx)
		Equal(fwdInOrder[idx], int(tree.First().Value), "First() (Iterator, remove)")
		visited++

		tree.Remove(node)
		tree.validate(t)
	}
	Equal(0, tree.Len(), "Len(): After iterating removal")
}

func (t *Tree[T]) validate(te *testing.T) {
	checkInvariants(te, t.root, nil)
}

func checkInvariants[T any](te *testing.T, node, parent *Node[T]) int {
	Equal := func(a, b any) {
		if !reflect.DeepEqual(a, b) {
			te.Error(a, "notEqual", b)
		}
	}
	if node == nil {
		return 0
	}

	// Validate the parent pointer.
	Equal(parent, node.parent)

	// Validate that the balance factor is -1, 0, 1.
	switch node.balance {
	case -1, 0, 1:
	default:
		te.Error(node.balance)
	}

	// Recursively derive the height of the left and right sub-trees.
	lHeight := checkInvariants(te, node.left, node)
	rHeight := checkInvariants(te, node.right, node)

	// Validate the AVL invariant and the balance factor.
	Equal(int(node.balance), rHeight-lHeight)
	if lHeight > rHeight {
		return lHeight + 1
	}
	return rHeight + 1
}

func BenchmarkAVLInsert(b *testing.B) {
	b.StopTimer()
	tree := New[int](func(a, b int) int {
		return a - b
	})
	for i := 0; i < 1e6; i++ {
		tree.Insert(int(i))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		v := (rand.Int() % 1e6) + 2e6
		tree.Insert(int(v))
	}
}

func BenchmarkAVLFind(b *testing.B) {
	b.StopTimer()
	tree := New[int](func(a, b int) int {
		return a - b
	})
	for i := 0; i < 1e6; i++ {
		tree.Insert(int(i))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		v := (rand.Int() % 1e6)
		tree.Find(int(v))
	}
}

func BenchmarkAVLDeleteLeft(b *testing.B) {
	b.StopTimer()
	tree := New[int](func(a, b int) int {
		return a - b
	})
	for i := 0; i < 5e6; i++ {
		tree.Insert(int(i))
	}
	b.StartTimer()
	it := tree.Iterator(Forward)
	nn := it.First()
	for i := 0; i < b.N; i++ {
		if nn = it.Get(); nn != nil {
			tree.Remove(nn)
			if it.Next() == nil {
				break
			}
		} else {
			break
		}
	}
}
