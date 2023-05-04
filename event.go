package event

import (
	"container/list"
	"sync"
)

type Subscriber[TEventData any] func(data TEventData)
type Unsubscribe func()

// Event represents a single event which can be subscribed and triggered
type Event[TEventData any] struct {
	mtx         sync.Mutex
	subscribers list.List
}

// On subscribes Subscriber to this event
func (e *Event[TEventData]) On(subscriber Subscriber[TEventData]) Unsubscribe {
	e.mtx.Lock()

	newSubscription := e.subscribers.PushBack(subscriber)

	e.mtx.Unlock()

	var once sync.Once

	return func() {
		once.Do(func() {
			e.mtx.Lock()

			e.subscribers.Remove(newSubscription)

			newSubscription = nil

			e.mtx.Unlock()
		})
	}
}

// Trigger calls all Subscriber subscribed to this event with eventData
func (e *Event[TEventData]) Trigger(eventData TEventData) {
	e.mtx.Lock()

	for i := e.subscribers.Front(); i != nil; i = i.Next() {
		e.mtx.Unlock()

		i.Value.(Subscriber[TEventData])(eventData)

		e.mtx.Lock()
	}

	e.mtx.Unlock()
}

// TriggerAsync is like the Trigger but calls every Subscriber in separate goroutine
func (e *Event[TEventData]) TriggerAsync(eventData TEventData) {
	wg := &sync.WaitGroup{}

	e.triggerAsync(wg, eventData)

	wg.Wait()
}

func (e *Event[TEventData]) triggerAsync(wg *sync.WaitGroup, eventData TEventData) {
	wg.Add(e.subscribers.Len())

	e.mtx.Lock()

	for i := e.subscribers.Front(); i != nil; i = i.Next() {
		go func(subscriber Subscriber[TEventData]) {
			subscriber(eventData)

			wg.Done()
		}(i.Value.(Subscriber[TEventData]))
	}

	e.mtx.Unlock()
}
