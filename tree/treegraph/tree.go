package treegraph

import (
	"tree/adjacencylist"
	"tree/graph"
	"tree/maybe"
	"tree/set"
)

type Tree[K comparable, V any] struct {
	maybeRoot maybe.Maybe[*Node[K, V]]
}

func (t *Tree[K, V]) Upsert(key K, value V) set.Set[K] {
	n, ok := t.maybeRoot.Get()
	if !ok {
		t.maybeRoot = maybe.Something(&Node[K, V]{[]*graph.Node[K, V]{}, key, value})
		return set.New(key)
	}

	s := n.Upsert(key, value, 3, set.Set[K]{})
	t.maybeRoot = maybe.Something(n)
	return s
}

func (t *Tree[K, V]) DeleteByKey(key interface{}) set.Set[K] {
	n, ok := t.maybeRoot.Get()
	if !ok {
		return set.Set[K]{}
	}

	maybeRoot, modifiedNodes := n.DeleteByKey(key, set.Set[K]{})
	t.maybeRoot = maybeRoot
	return modifiedNodes
}

func (t Tree[K, V]) Find(key K) (maybe.Maybe[V], bool) {
	n, ok := t.maybeRoot.Get()
	if !ok {
		return maybe.Nothing[V](), false
	}
	return (*graph.Node[K, V])(n).Find(key)
}

func (t Tree[K, V]) Has(key K) bool {
	n, ok := t.maybeRoot.Get()
	if !ok {
		return false
	}
	return (*graph.Node[K, V])(n).Has(key)
}

func (t Tree[K, V]) AdjacencyList() adjacencylist.AdjacencyList[K, V] {
	n, ok := t.maybeRoot.Get()
	if !ok {
		return adjacencylist.AdjacencyList[K, V]{}
	}
	return (*graph.Node[K, V])(n).AdjacencyList(set.Set[K]{})
}

func (t Tree[K, V]) IsEmpty() bool {
	_, ok := t.maybeRoot.Get()
	return !ok
}
