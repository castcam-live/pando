package maybe

type Maybe[V any] struct {
	node   V
	exists bool
}

func Something[V any](node V) Maybe[V] {
	return Maybe[V]{
		node:   node,
		exists: true,
	}
}

func Nothing[V any]() Maybe[V] {
	var noop V
	return Maybe[V]{
		node:   noop,
		exists: false,
	}
}

func (m Maybe[V]) Get() (V, bool) {
	return m.node, m.exists
}
