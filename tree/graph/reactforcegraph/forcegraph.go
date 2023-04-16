package reactforcegraph

import (
	"encoding/json"
	"tree/graph/adjacencylist"
)

type ReactForceGraphMarshaler[K comparable, V any] adjacencylist.AdjacencyList[K, V]

var _ json.Marshaler = &ReactForceGraphMarshaler[string, int]{}

type ForceGraphNode[K comparable] struct {
	ID K
}

type ForceGraphLink[K comparable] struct {
	Target K `json:"target"`
	Source K `json:"source"`
}

type forceGraph[K comparable] struct {
	Nodes []ForceGraphNode[K] `json:"nodes"`
	Links []ForceGraphLink[K] `json:"links"`
}

func (s ReactForceGraphMarshaler[K, V]) MarshalJSON() ([]byte, error) {
	nodes := []ForceGraphNode[K]{}
	links := []ForceGraphLink[K]{}
	for k, list := range s {
		nodes = append(nodes, ForceGraphNode[K]{ID: k})

		for link := range list.Neighbors {
			links = append(links, ForceGraphLink[K]{k, link})
		}
	}

	graph := forceGraph[K]{nodes, links}
	return json.Marshal(graph)
}
