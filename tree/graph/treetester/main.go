// This is a useless

package main

import (
	"encoding/json"
	"fmt"
	"tree/graph/graph"
	"tree/graph/set"
	insertiontree "tree/graph/treegraph"
	"tree/graph/treegraph/nodesandedges"
)

func main() {
	tree := insertiontree.Node[string, byte]{Neighbors: []*graph.Node[string, byte]{}, Key: "cool", Value: 1}

	tree.Upsert("foo", 2, 3, set.Set[string]{})
	tree.Upsert("bar", 2, 3, set.Set[string]{})
	tree.Upsert("baz", 2, 3, set.Set[string]{})
	tree.Upsert("foobar", 2, 3, set.Set[string]{})
	tree.Upsert("widgets", 2, 3, set.Set[string]{})
	tree.Upsert("gadgets", 2, 3, set.Set[string]{})
	tree.Upsert("hello", 2, 3, set.Set[string]{})
	tree.Upsert("world", 2, 3, set.Set[string]{})
	tree.Upsert("sweet", 2, 3, set.Set[string]{})

	fgraph := nodesandedges.NodesAndEdges[string, byte](graph.Node[string, byte](tree).AdjacencyList(set.Set[string]{}))

	b, err := json.Marshal(fgraph)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
