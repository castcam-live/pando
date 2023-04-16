package treemanager

import (
	"sync"
	"tree/graph/treegraph"
	"tree/graph/treemanager/listeners"
	"tree/graph/treemanager/safetree"
)

type treeManager[K comparable, V any] struct {
	mut       *sync.RWMutex
	trees     map[string]*safetree.SafeTree[K, V]
	listeners listeners.KeyedListeners
}

func NewTreeManager[K comparable, V any]() treeManager[K, V] {
	managerMut := &sync.RWMutex{}
	return treeManager[K, V]{
		mut:   managerMut,
		trees: make(map[string]*safetree.SafeTree[K, V]),
	}
}

func (t *treeManager[K, V]) GetTree(id string) *safetree.SafeTree[K, V] {
	t.mut.Lock()
	defer t.mut.Unlock()

	tree, ok := t.trees[id]
	if !ok {
		newTree := safetree.New[K, V]()
		t.trees[id] = &newTree
		tree = &newTree
	}

	return tree
}

func (t *treeManager[K, V]) Upsert(treeId string, nodeId K, p V) {
	t.mut.Lock()
	defer t.mut.Unlock()
	tree := t.GetTree(treeId)

	changedNodes := tree.Upsert(nodeId, p)
	t.listeners.EmitEvent(treeId, changedNodes)
}

func (t *treeManager[K, V]) GetNeighborOfNode(treeId string, nodeId K) ([]treegraph.Pair[K, V], bool) {
	t.mut.Lock()
	defer t.mut.Unlock()
	tree := t.GetTree(treeId)

	neighbor, ok := tree.GetNeighborOfNode(nodeId)
	if !ok {
		return nil, false
	}

	return neighbor, true
}

func (t *treeManager[K, V]) DeleteNode(treeId, nodeId string) {
	t.mut.Lock()
	defer t.mut.Unlock()
	tree := t.GetTree(treeId)

	changedNodes := tree.DeleteByKey(nodeId)
	if tree.IsEmpty() {
		delete(t.trees, treeId)
	}

	t.listeners.EmitEvent(treeId, changedNodes)
}

func (t *treeManager[K, V]) RegisterChangeListener(
	treeId interface{},
) <-chan interface{} {
	return t.listeners.RegisterListener(treeId)
}

func (t *treeManager[K, V]) UnregisterChangeListener(
	treeId interface{},
	listener <-chan interface{},
) {
	t.listeners.UnregisterListener(treeId, listener)
}
