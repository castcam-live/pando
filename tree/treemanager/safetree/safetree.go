package safetree

import (
	"sync"
	"tree/adjacencylist"
	"tree/maybe"
	"tree/set"
	"tree/treegraph"
)

type SafeTree[K comparable, V any] struct {
	mut  *sync.RWMutex
	tree treegraph.Tree[K, V]
}

func New[K comparable, V any]() SafeTree[K, V] {
	mut := &sync.RWMutex{}
	return SafeTree[K, V]{mut, treegraph.Tree[K, V]{}}
}

func (t *SafeTree[K, V]) Upsert(key K, value V) set.Set[K] {
	t.mut.Lock()
	defer t.mut.Unlock()
	return t.tree.Upsert(key, value)
}

func (t *SafeTree[K, V]) DeleteByKey(key interface{}) set.Set[K] {
	t.mut.Lock()
	defer t.mut.Unlock()
	return t.tree.DeleteByKey(key)
}

func (t SafeTree[K, V]) Find(key K) (maybe.Maybe[V], bool) {
	t.mut.RLock()
	defer t.mut.RUnlock()
	return t.tree.Find(key)
}

func (t SafeTree[K, V]) GetNeighborOfNode(key K) ([]treegraph.Pair[K, V], bool) {
	t.mut.RLock()
	defer t.mut.RUnlock()
	return t.tree.GetNeighborsOfNode(key)
}

func (t SafeTree[K, V]) Has(key K) bool {
	t.mut.RLock()
	defer t.mut.RUnlock()
	return t.tree.Has(key)
}

func (t SafeTree[K, V]) AdjacencyList() adjacencylist.AdjacencyList[K, V] {
	t.mut.RLock()
	defer t.mut.RUnlock()
	return t.tree.AdjacencyList()
}

func (t SafeTree[K, V]) IsEmpty() bool {
	t.mut.RLock()
	defer t.mut.RUnlock()
	return t.tree.IsEmpty()
}
