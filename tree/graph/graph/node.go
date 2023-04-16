package graph

import (
	"tree/graph/adjacencylist"
	"tree/graph/set"
)

// Node represents a node in a graph.
//
// This structure does not care the least about what the graph represents.
// It could be a directed tree. Directed acyclic lattice. Undirected tree.
// Full mesh, etc. It should not matter. This structure does not care.
//
// Other entities can, however, use this structure to enforce its use in a
// particular graph configuration, but this structure itself is not responsible
// for anything, other than to provide basic primitive that graph theorists may
// depend on
type Node[K comparable, V any] struct {
	Neighbors []*Node[K, V]
	Key       K
	Value     V
}

// Pair representing a tuple containing the key and value that would typically
// be held by a node
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

// Cleave cleaves the node, and returns its neighbors, and a set of keys
// associated with all modified nodes.
func (n *Node[K, V]) Cleave() ([]*Node[K, V], set.Set[K]) {
	// Trivially, all neighbors will be modified, so grab their key set
	modified := n.GetNeighborKeys()

	// Let's not forget the newly cleaved node as well
	modified.Add(n.Key)

	// Grab the neigbors
	neighbors := n.Neighbors

	for _, neighbor := range neighbors {
		// Remove references of the cleaved node from all neighboring nodes
		neighbor.Neighbors =
			ExcludeNodesByKeys(neighbor.Neighbors, set.New(n.Key))
	}

	// The cleaved node is now a lone node
	n.Neighbors = []*Node[K, V]{}

	// Return the neighbors and the set of modified nodes
	return neighbors, modified
}

// Interject will take a slice of distinct graphs (each pointed to by an
// orginating node), and attach them
func (n *Node[K, V]) Interject(newNeighbors []*Node[K, V]) set.Set[K] {
	s := set.New(n.Key)
	for _, neighbor := range newNeighbors {
		s.Add(neighbor.Key)
		neighbor.Neighbors = append(n.Neighbors, n)
		n.Neighbors = append(n.Neighbors, neighbor)
	}
	return s
}

// Gets the keys of the neighboring nodes
func (n Node[K, V]) GetNeighborKeys() set.Set[K] {
	keys := set.Set[K]{}

	for _, neighbor := range n.Neighbors {
		keys.Add(neighbor.Key)
	}

	return keys
}

// AdjacencyList performs a breadth-first-search of the node and creates an
// adjacency list of all neighboring nodes of all nodes that the BFS encountered
func (n Node[K, V]) AdjacencyList(visited set.Set[K]) adjacencylist.AdjacencyList[K, V] {
	list := adjacencylist.AdjacencyList[K, V]{}
	list[n.Key] = adjacencylist.AdjacencyListNode[K, V]{
		Value:     n.Value,
		Neighbors: n.GetNeighborKeys(),
	}

	for _, neighbor := range n.Neighbors {
		if !visited.Has(neighbor.Key) {
			list = list.Union(neighbor.AdjacencyList(visited.Union(set.New(n.Key))))
		}
	}

	return list
}

// Traverse iterates through all nodes in the graph, on a depth-first-search
// basis, ensuring to avoid traversing the same node more than once
func (n *Node[K, V]) Traverse(visited set.Set[K]) <-chan *Node[K, V] {
	visited.Add(n.Key)

	c := make(chan *Node[K, V], 3)

	go func() {
		// Emit the current node to the listener
		c <- n

		// Iterate through all the neighbors
		for _, neighbor := range n.Neighbors {

			// Check to see if we have not visited the neighbor yet
			if !visited.Has(neighbor.Key) {

				// If not, add it to our visited list
				visited.Add(neighbor.Key)

				// Do a traversal on the neighbouring subtree
				for child := range neighbor.Traverse(visited) {
					c <- child
				}
			}
		}

		close(c)
	}()

	return c
}

// ToSlice gets all key/value pairs in the graph, by traversing each node via
// neighboring nodes on a DFS basis
func (n Node[K, V]) ToSlice() []Pair[K, V] {
	result := []Pair[K, V]{}

	for c := range n.Traverse(set.Set[K]{}) {
		result = append(result, Pair[K, V]{Key: c.Key, Value: c.Value})
	}

	return result
}

// Cardinality gets the count of all reachable nodes, through a DFS traversal
func (n Node[K, V]) Cardinality() int {
	return len(n.ToSlice())
}

// Find gets the value associated with the supplied key
func (n Node[K, V]) Find(key K) (*Node[K, V], bool) {
	for node := range n.Traverse(set.Set[K]{}) {
		if node.Key == key {
			return node, true
		}
	}

	return nil, false
}

// Has determines whether a node with the supplied key exists in the graph
func (n Node[K, V]) Has(key K) bool {
	_, ok := n.Find(key)
	return ok
}

func (n Node[K, V]) GetMap() map[interface{}]interface{} {
	m := map[interface{}]interface{}{}

	for c := range n.Traverse(set.Set[K]{}) {
		m[c.Key] = c.Value
	}

	return m
}
