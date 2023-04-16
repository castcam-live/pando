package graph

import (
	"testing"
	"tree/graph/set"
)

type stringSet map[string]bool

func newStringSet(strs ...string) stringSet {
	s := stringSet{}
	for _, str := range strs {
		s[str] = true
	}
	return s
}

func (s stringSet) add(str string) {
	s[str] = true
}

func (s stringSet) equal(b stringSet) bool {
	if len(s) != len(b) {
		return false
	}

	for k, v := range s {
		if !b[k] || !v {
			return false
		}
	}

	return true
}

func TestExcludeNodesByKeys(t *testing.T) {
	nodes := []*Node[string, string]{
		{[]*Node[string, string]{}, "hello", "1"},
		{[]*Node[string, string]{}, "world", "1"},
		{[]*Node[string, string]{}, "foo", "1"},
		{[]*Node[string, string]{}, "bar", "1"},
		{[]*Node[string, string]{}, "baz", "1"},
		{[]*Node[string, string]{}, "foobar", "1"},
		{[]*Node[string, string]{}, "widgets", "1"},
		{[]*Node[string, string]{}, "gadgets", "1"},
	}

	expectedList := newStringSet("hello", "world", "bar", "baz", "widgets")

	toExclude := set.New("foo", "foobar", "gadgets")

	withExclusion := ExcludeNodesByKeys(nodes, toExclude)

	newList := stringSet{}
	for _, n := range withExclusion {
		newList.add(n.Key)
	}

	if !newList.equal(expectedList) {
		t.Logf("expected %v\nActual %v", expectedList, newList)
		t.Fail()
	}
}

func TestNodesHaveKeys(t *testing.T) {
	nodes := []*Node[string, string]{
		{[]*Node[string, string]{}, "hello", "1"},
		{[]*Node[string, string]{}, "world", "1"},
		{[]*Node[string, string]{}, "foo", "1"},
		{[]*Node[string, string]{}, "bar", "1"},
		{[]*Node[string, string]{}, "baz", "1"},
		{[]*Node[string, string]{}, "foobar", "1"},
		{[]*Node[string, string]{}, "widgets", "1"},
		{[]*Node[string, string]{}, "gadgets", "1"},
	}

	if !NodesHaveKeys(nodes, set.New("bar")) {
		t.Fail()
	}

	if !NodesHaveKeys(nodes, set.New("hello")) {
		t.Fail()
	}

	if NodesHaveKeys(nodes, set.New("sweet")) {
		t.Fail()
	}
}
