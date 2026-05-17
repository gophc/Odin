package util

import (
	"testing"
)

func TestRBTree(t *testing.T) {
	tree := &RBTree[string]{}

	t.Run("Insert and Size", func(t *testing.T) {
		tree.Insert(5, "five")
		tree.Insert(3, "three")
		tree.Insert(7, "seven")

		if tree.Size != 3 {
			t.Errorf("expected Size to be 3, got %v", tree.Size)
		}

		if val := tree.Root.Value; val != "five" {
			t.Errorf("expected Root Value to be 'five', got %v", val)
		}

		if val := tree.Root.Left.Value; val != "three" {
			t.Errorf("expected Left Value to be 'three', got %v", val)
		}

		if val := tree.Root.Right.Value; val != "seven" {
			t.Errorf("expected Right Value to be 'seven', got %v", val)
		}
	})
}

func buildRBTree[T any](vals []*RBNode[T], f func(*RBTree[T], *RBNode[T])) *RBTree[T] {
	tree := NewRBTree[T]()
	for _, val := range vals {
		f(tree, val)
	}
	return tree
}

func buildRBTreeMap[T any](vals []*RBNode[T]) map[string]*RBTree[T] {
	m := make(map[string]*RBTree[T])
	m["Insert"] = buildRBTree(vals, func(tree *RBTree[T], val *RBNode[T]) {
		tree.Insert(val.Key, val.Value)
	})
	m["InsertNode"] = buildRBTree(vals, func(tree *RBTree[T], val *RBNode[T]) {
		tree.InsertNode(val)
	})
	return m
}

func TestRBTreeBase(t *testing.T) {
	vals := []*RBNode[string]{
		NewRBNode(5, "five"),
		NewRBNode(3, "three"),
		NewRBNode(7, "seven"),
	}

	for key, tree := range buildRBTreeMap(vals) {
		t.Run(key, func(t *testing.T) {
			if tree.Size != 3 {
				t.Errorf("expected Size to be 3, got %v", tree.Size)
			}

			if val := tree.Root.Value; val != "five" {
				t.Errorf("expected Root Value to be 'five', got %v", val)
			}

			if val := tree.Root.Left.Value; val != "three" {
				t.Errorf("expected Left Value to be 'three', got %v", val)
			}

			if val := tree.Root.Right.Value; val != "seven" {
				t.Errorf("expected Right Value to be 'seven', got %v", val)
			}
		})
	}
}

func TestRBTreeOrder(t *testing.T) {
	vals := []*RBNode[string]{
		NewRBNode(3, "three"),
		NewRBNode(1, "one"),
		NewRBNode(2, "two"),
		NewRBNode(4, "four"),
	}

	for key, tree := range buildRBTreeMap(vals) {
		t.Run(key, func(t *testing.T) {
			nodes := tree.InOrder()
			if len(nodes) != 4 {
				t.Errorf("Expected 4 nodes, got %d", len(nodes))
			}

			expectedKeys := []int{1, 2, 3, 4}
			for i, node := range nodes {
				if node.Key != uintptr(expectedKeys[i]) {
					t.Errorf("InOrder Expected key %d, got %d", expectedKeys[i], node.Key)
				}
			}

			nodes2 := tree.ReverseInOrder()
			expectedKeys2 := []int{4, 3, 2, 1}
			for i, node := range nodes2 {
				if node.Key != uintptr(expectedKeys2[i]) {
					t.Errorf("ReverseInOrder Expected key %d, got %d", expectedKeys2[i], node.Key)
				}
			}
		})
	}
}

func TestInsertAndGet(t *testing.T) {
	vals := []*RBNode[string]{
		NewRBNode(3, "three"),
		NewRBNode(1, "one"),
		NewRBNode(2, "two"),
		NewRBNode(4, "four"),
	}

	for key, tree := range buildRBTreeMap(vals) {
		t.Run(key, func(t *testing.T) {
			if value := tree.Get(1); value != "one" {
				t.Errorf("Expected 'one', got %v", value)
			}

			if value := tree.Get(2); value != "two" {
				t.Errorf("Expected 'two', got %v", value)
			}

			if value := tree.Get(3); value != "three" {
				t.Errorf("Expected 'three', got %v", value)
			}

			if value := tree.Get(4); value != "four" {
				t.Errorf("Expected 'four', got %v", value)
			}

			if value := tree.Get(5); value != "" {
				t.Errorf("Expected nil, got %v", value)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	vals := []*RBNode[string]{
		NewRBNode(3, "three"),
		NewRBNode(1, "one"),
		NewRBNode(2, "two"),
		NewRBNode(4, "four"),
	}

	for key, tree := range buildRBTreeMap(vals) {
		t.Run(key, func(t *testing.T) {
			tree.Delete(1)
			if value := tree.Get(1); value != "" {
				t.Errorf("Expected nil, got %v", value)
			}

			tree.Delete(2)
			if value := tree.Get(2); value != "" {
				t.Errorf("Expected nil, got %v", value)
			}

			tree.Delete(3)
			if value := tree.Get(3); value != "" {
				t.Errorf("Expected nil, got %v", value)
			}

			tree.Delete(4)
			if value := tree.Get(4); value != "" {
				t.Errorf("Expected nil, got %v", value)
			}
		})
	}
}

func FuzzInsert(f *testing.F) {
	f.Add(3, "three")
	f.Add(1, "one")
	f.Add(2, "two")
	f.Add(4, "four")

	f.Fuzz(func(t *testing.T, key int, value string) {
		tree := NewRBTree[string]()
		tree.Insert(uintptr(key), value)
		if val := tree.Get(uintptr(key)); val != value {
			t.Errorf("Expected %v, got %v", value, val)
		}
	})
}

func FuzzDelete(f *testing.F) {
	f.Add(3, "three")
	f.Add(1, "one")
	f.Add(2, "two")
	f.Add(4, "four")

	f.Fuzz(func(t *testing.T, key int, value string) {
		tree := NewRBTree[string]()
		tree.Insert(uintptr(key), value)
		tree.Delete(uintptr(key))
		if val := tree.Get(uintptr(key)); val != "" {
			t.Errorf("Expected nil, got %v", val)
		}
	})
}
