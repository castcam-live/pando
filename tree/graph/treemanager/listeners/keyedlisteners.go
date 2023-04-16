package listeners

import "sync"

type KeyedListeners struct {
	mut       *sync.RWMutex
	listeners map[interface{}]*Listeners
}

func NewKeyedListeners() KeyedListeners {
	mut := &sync.RWMutex{}
	return KeyedListeners{
		mut:       mut,
		listeners: make(map[interface{}]*Listeners),
	}
}

func (k *KeyedListeners) RegisterListener(key interface{}) <-chan interface{} {
	k.mut.Lock()
	defer k.mut.Unlock()

	k.initializeListeners()

	listeners, ok := k.listeners[key]
	if !ok {
		listeners = &Listeners{}
		k.listeners[key] = listeners
	}
	return listeners.RegisterListener()
}

func (k *KeyedListeners) initializeListeners() {
	if k.listeners == nil {
		k.listeners = make(map[interface{}]*Listeners)
	}
}

func (k *KeyedListeners) UnregisterListener(key interface{}, c <-chan interface{}) {
	k.mut.Lock()
	defer k.mut.Unlock()

	k.initializeListeners()

	listeners, ok := k.listeners[key]
	if !ok {
		return
	}

	listeners.UnregisterListener(c)
}

func (k *KeyedListeners) EmitEvent(key, d interface{}) {
	k.mut.RLock()
	defer k.mut.RUnlock()

	listeners, ok := k.listeners[key]
	if ok {
		listeners.EmitEvent(d)
	}
}

func (k *KeyedListeners) EmitEventToAll(d interface{}) {
	k.mut.RLock()
	defer k.mut.RUnlock()

	for _, value := range k.listeners {
		value.EmitEvent(d)
	}
}

type Pair struct {
	key       interface{}
	listeners *Listeners
}

func (k *KeyedListeners) Iterate() <-chan Pair {
	c := make(chan Pair)

	go func() {
		k.mut.RLock()
		defer k.mut.RUnlock()

		for key, value := range k.listeners {
			c <- Pair{key, value}
		}
		close(c)
	}()

	return c
}

func (k *KeyedListeners) Get(key interface{}) (*Listeners, bool) {
	k.mut.RLock()
	defer k.mut.RUnlock()

	l, ok := k.listeners[key]
	return l, ok
}
