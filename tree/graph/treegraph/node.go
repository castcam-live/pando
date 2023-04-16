package treegraph

import (
	"tree/graph/graph"
	"tree/graph/maybe"
	"tree/graph/set"
)

// Node represents a node in a undirected tree.
//
// This type contains methods specifically intended to traverse and manipulate
// such a tree graph
type Node[K comparable, V any] graph.Node[K, V]

// Upsert takes the key and value, and upserts it into the graph. That is, if
// a node with the key exists, the value of the node gets replaced by the
// supplied value; otherwise, a new node will be created as a leaf onto the
// shortest subtree
func (n *Node[K, V]) Upsert(
	key K, value V,
	maxNeighbors int,
	visited set.Set[K],
) set.Set[K] {
	if n.Key == key {
		n.Value = value
		return set.New(key)
	}

	maybeSubTree := n.ShortestSubTree(set.Set[K]{})

	subTree, ok := maybeSubTree.Get()

	if !ok || len(n.Neighbors) < maxNeighbors {
		parent := (*graph.Node[K, V])(n)
		newNode := &graph.Node[K, V]{
			Neighbors: []*graph.Node[K, V]{parent},
			Key:       key,
			Value:     value,
		}
		n.Neighbors = append(
			n.Neighbors,
			newNode,
		)
		return set.New(n.Key, key)
	}

	return subTree.Upsert(key, value, maxNeighbors, visited.Union(set.New(n.Key)))
}

// LongestPath returns an ordered slice of nodes that will represent the longest
// path originating from the node.
//
// This algorithm is not likely to work with any arbitrary graph; only with
// trees
func (n *Node[K, V]) LongestPath(visited set.Set[K]) []*Node[K, V] {
	newVisited := visited.Union(set.New(n.Key))

	withoutVisited := graph.ExcludeNodesByKeys(n.Neighbors, newVisited)

	neighbouringSubTrees := [][]*Node[K, V]{}
	for _, neighbor := range withoutVisited {
		neighbouringSubTrees =
			append(neighbouringSubTrees, (*Node[K, V])(neighbor).LongestPath(newVisited))
	}

	if len(neighbouringSubTrees) <= 0 {
		return []*Node[K, V]{n}
	}

	longest := neighbouringSubTrees[0]
	for _, neighbor := range neighbouringSubTrees {
		if len(neighbor) > len(longest) {
			longest = neighbor
		}
	}

	return longest
}

// GetLeafiestNode gets the leaf node of the longest path
//
// This algorithm is not likely to work with any arbitrary graph; only with
// trees
func (n *Node[K, V]) GetLeafiestNode(visited set.Set[K]) *Node[K, V] {
	chain := n.LongestPath(visited)
	if len(chain) <= 0 {
		return n
	}

	return chain[len(chain)-1]
}

// CleaveLeafiestNode removes the leafiest node from the graph
//
// This algorithm is not likely to work with any arbitrary graph; only with
// trees
func (n *Node[K, V]) CleaveLeafiestNode(visited set.Set[K]) (*Node[K, V], set.Set[K]) {
	leaf := n.GetLeafiestNode(visited)
	_, modified := (*graph.Node[K, V])(n).Cleave()
	return leaf, modified
}

// DeleteByKey deletes the node in the graph
func (n *Node[K, V]) DeleteByKey(
	key interface{},
	visited set.Set[K],
) (maybe.Maybe[*Node[K, V]], set.Set[K]) {

	// This means we found our node to delete
	if n.Key == key {
		// Grab the leafiest node, and cleave it off of the tree
		leaf, modified := n.CleaveLeafiestNode(visited)

		// If the leaf node is the node to delete, then just return null
		if leaf.Key == n.Key {
			return maybe.Nothing[*Node[K, V]](), modified.Union(set.New(leaf.Key))
		}

		// Cleave this node, and interject the leaf node in its place
		neighbors, cleavedModified := (*graph.Node[K, V])(n).Cleave()
		interjectedModified := (*graph.Node[K, V])(leaf).Interject(neighbors)

		return maybe.Something(leaf), modified.Union(cleavedModified).Union(interjectedModified)
	}

	// Start from a black slate of neighbors
	newNeighbors := []*graph.Node[K, V]{}

	modifiedList := set.Set[K]{}

	// Iterate through the current set of neighbors, performing a depth-first
	// traversal for deletion
	for _, neighbor := range n.Neighbors {
		// Remember to check if the neighboring node has been visited already
		if !visited.Has(neighbor.Key) {
			// Perform DFS on neighboring node, while also registering current node as
			// visited node
			withDeletionMaybe, modifiedKeysList :=
				(*Node[K, V])(neighbor).DeleteByKey(key, visited.Union(set.New(n.Key)))

			// Append all modified keys
			modifiedList = modifiedList.Union((modifiedKeysList))

			if withDeletion, ok := withDeletionMaybe.Get(); ok {
				newNeighbors =
					graph.AddNeighbor(newNeighbors, (*graph.Node[K, V])(withDeletion))
			} else {
				modifiedList.Add(n.Key)
			}
		} else {
			modifiedList.Add(n.Key)
			newNeighbors = graph.AddNeighbor(newNeighbors, neighbor)
		}
	}

	n.Neighbors = newNeighbors

	return maybe.Something(n), modifiedList
}

// ShortestPath returns an ordered slice of nodes representing the shortest path
// in the graph
func (n *Node[K, V]) ShortestPath(visited set.Set[K]) []*Node[K, V] {
	newVisited := visited.Union(set.New(n.Key))

	// Get all neighbors that have yet to be visited
	withoutVisited := graph.ExcludeNodesByKeys(n.Neighbors, newVisited)

	// Get the shortest path of all sub trees
	neighboringSubTrees := [][]*Node[K, V]{}
	for _, neighbor := range withoutVisited {
		neighboringSubTrees =
			append(neighboringSubTrees, (*Node[K, V])(neighbor).ShortestPath(newVisited))
	}

	if len(neighboringSubTrees) <= 0 {
		return []*Node[K, V]{n}
	}

	shortest := neighboringSubTrees[0]
	for _, neighbor := range neighboringSubTrees {
		if len(neighbor) < len(shortest) {
			shortest = neighbor
		}
	}

	return append([]*Node[K, V]{n}, shortest...)
}

// ShortestSubTree returns the node that represents the shortest sub tree of the
// node in question
func (n Node[K, V]) ShortestSubTree(visited set.Set[K]) maybe.Maybe[*Node[K, V]] {
	if len(graph.ExcludeNodesByKeys(n.Neighbors, visited)) <= 0 {
		return maybe.Nothing[*Node[K, V]]()
	}

	next := n.ShortestPath(visited)[1]

	for _, neighbor := range n.Neighbors {
		if neighbor.Key == next.Key {
			return maybe.Something((*Node[K, V])(neighbor))
		}
	}

	return maybe.Nothing[*Node[K, V]]()
}
