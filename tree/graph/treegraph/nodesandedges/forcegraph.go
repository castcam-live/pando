package nodesandedges

import (
	"encoding/json"
	"tree/graph/adjacencylist"
)

type NodesAndEdges[K comparable, V any] adjacencylist.AdjacencyList[K, V]

var _ json.Marshaler = &NodesAndEdges[string, int]{}

type Node[K comparable] struct {
	ID K
}

type Link[K comparable] struct {
	Target K `json:"target"`
	Source K `json:"source"`
}

type graph[K comparable] struct {
	Nodes []Node[K] `json:"nodes"`
	Links []Link[K] `json:"links"`
}

func (s NodesAndEdges[K, V]) MarshalJSON() ([]byte, error) {
	nodes := []Node[K]{}
	links := []Link[K]{}
	for k, list := range s {
		nodes = append(nodes, Node[K]{ID: k})

		for link := range list.Neighbors {
			links = append(links, Link[K]{k, link})
		}
	}

	graph := graph[K]{nodes, links}
	return json.Marshal(graph)
}
