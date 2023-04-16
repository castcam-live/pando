package graph

import "tree/graph/set"

// ExcludeNodesByKeys given a slice of nodes, return a slice of nodes that does
// not have nodes whose keys are in the provided key set
func ExcludeNodesByKeys[K comparable, V any](nodes []*Node[K, V], keys set.Set[K]) []*Node[K, V] {
	newList := []*Node[K, V]{}

	for _, node := range nodes {
		if !keys.Has(node.Key) {
			newList = append(newList, node)
		}
	}

	return newList
}

// NodesHaveKeys determines if any of the nodes in the nodes slice has keys as
// determined by the set of keys in the keys set
func NodesHaveKeys[K comparable, V any](nodes []*Node[K, V], keys set.Set[K]) bool {
	for _, node := range nodes {
		if keys.Has(node.Key) {
			return true
		}
	}
	return false
}

// AddNeighbor is a helper function for adding a node to a list of nodes.
//
// The difference between this function and `append` is that this method will
// only add the node if the key does not exist in the slice. Otherwise, the
// function call becomes a no-op
func AddNeighbor[K comparable, V any](nodes []*Node[K, V], node *Node[K, V]) []*Node[K, V] {
	if !NodesHaveKeys(nodes, set.New(node.Key)) {
		return append(nodes, node)
	}
	return nodes
}
