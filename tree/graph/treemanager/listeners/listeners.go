package listeners

import (
	"sync"
)

type Listeners struct {
	mut       sync.RWMutex
	listeners [](chan interface{})
}

func (l *Listeners) RegisterListener() <-chan interface{} {
	l.mut.Lock()
	defer l.mut.Unlock()
	c := make(chan interface{})
	l.listeners = append(l.listeners, c)
	return c
}

func (l *Listeners) UnregisterListener(c <-chan interface{}) {
	l.mut.Lock()
	defer l.mut.Unlock()

	listeners := make([](chan interface{}), 0)

	for _, listener := range l.listeners {
		if listener != c {
			listeners = append(listeners, listener)
		} else {
			close(listener)
		}
	}
	l.listeners = listeners
}

func (l *Listeners) EmitEvent(d interface{}) {
	l.mut.RLock()
	defer l.mut.RUnlock()

	for _, listener := range l.listeners {
		go func(l chan interface{}) {
			l <- d
		}(listener)
	}
}
