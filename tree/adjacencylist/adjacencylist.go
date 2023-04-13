package adjacencylist

import (
	"fmt"
	"tree/maybe"
	"tree/set"
)

// AdjacencyListNode represents a pairing from source node to a list of target
// nodes
type AdjacencyListNode[K comparable, V any] struct {
	Value     V
	Neighbors set.Set[K]
}

// UnionLinks creates a new AdjacencyListNode with links added to the pre-
// existing set of links
func (a AdjacencyListNode[K, V]) UnionLinks(links set.Set[K]) AdjacencyListNode[K, V] {
	return AdjacencyListNode[K, V]{Value: a.Value, Neighbors: links.Union(a.Neighbors)}
}

// UnionLinks creates a new AdjacencyListNode with the link added to the pre-
// existing set of links
func (a AdjacencyListNode[K, V]) AddLink(link K) AdjacencyListNode[K, V] {
	return AdjacencyListNode[K, V]{Value: a.Value, Neighbors: a.Neighbors.Union(set.New(link))}
}

// AdjacencyList represents a mapping of all nodes to other links in the network
type AdjacencyList[K comparable, V any] map[K]AdjacencyListNode[K, V]

// Sets an adjacency link to the AdjacencyList
func (a *AdjacencyList[K, V]) SetLink(key K, value V) {
	link, ok := (*a)[key]
	if !ok {
		(*a)[key] = AdjacencyListNode[K, V]{value, set.Set[K]{}}
		return
	}
	(*a)[key] = AdjacencyListNode[K, V]{value, link.Neighbors}
}

func (a *AdjacencyList[K, V]) AddLinks(key K, link set.Set[K], defaultValue V) {
	l, ok := (*a)[key]
	if !ok {
		(*a)[key] = AdjacencyListNode[K, V]{defaultValue, link}
	} else {
		(*a)[key] = AdjacencyListNode[K, V]{l.Value, l.Neighbors.Union(link)}
	}
}

// GetReversed gets the graph represented by the adjacencylist, but with the edges
// reversed
func (a AdjacencyList[K, V]) GetReversed() AdjacencyList[K, V] {
	newList := AdjacencyList[K, V]{}

	// Iterate through the original list of ndoes
	for key, node := range a {

		// Then, for every link, point it to key
		for link := range node.Neighbors {
			newList.AddLinks(link, set.New(key), node.Value)
		}

	}

	return newList
}

// Union combines two adjacency lists into a single graph. Especially useful if
// combined with `Reversed` to drive an bidirectional (undirected) graph.
func (a AdjacencyList[K, V]) Union(b AdjacencyList[K, V]) AdjacencyList[K, V] {
	newList := AdjacencyList[K, V]{}

	for key, node := range a {
		newList.AddLinks(key, node.Neighbors, node.Value)
	}

	for key, node := range b {
		newList.AddLinks(key, node.Neighbors, node.Value)
	}

	return newList
}

// Equal tests the equality of two adjacency lists. The graph not only must be
// homomorphic from each other, but they must key-by-key (e.g. it is not
// sufficient for two graphs to have the same shape, but each node must have
// the same keys, and must direct to the same keys)
func (a AdjacencyList[K, V]) Equal(a1 AdjacencyList[K, V]) bool {
	if len(a) != len(a1) {
		return false
	}

	for key, n := range a {
		node, ok := a1[key]
		if !ok {
			return false
		}

		if !n.Neighbors.Equals(node.Neighbors) {
			return false
		}
	}

	return true
}

// GetKeys gets the list of keys in the graph
func (a AdjacencyList[K, V]) GetKeys() set.Set[K] {
	result := set.Set[K]{}
	for k := range a {
		result.Add(k)
	}
	return result
}

func (a AdjacencyList[K, V]) GetAnyKey() maybe.Maybe[K] {
	for k := range a {
		return maybe.Something(k)
	}

	return maybe.Nothing[K]()
}

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

func (a AdjacencyList[K, V]) Traverse(currentNode K, visited set.Set[K]) <-chan Pair[K, V] {
	c := make(chan Pair[K, V])

	go func() {
		defer close(c)

		node, ok := a[currentNode]
		if !ok {
			return
		}

		visited.Add(currentNode)
		c <- Pair[K, V]{currentNode, node.Value}

		// Iterate through each of the keys of the neighboring nodes
		for neighborKey := range node.Neighbors {
			if !visited.Has(neighborKey) {
				next := a.Traverse(neighborKey, visited)

				for pair := range next {
					c <- pair
				}
			}
		}
	}()

	return c
}

func GetKeysFromTraversal[K comparable, V any](traversal <-chan Pair[K, V]) set.Set[K] {
	s := set.Set[K]{}

	for pair := range traversal {
		s.Add(pair.Key)
	}

	fmt.Println("The visited set", s)

	return s
}
