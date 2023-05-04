package event

import (
	"sync"
)

// MultiSubscriber is a subscriber to multiple events
type MultiSubscriber[TEventData any] func(eventName string, data TEventData)

// A Subscribable represents an interface which can be used for disable triggering outside some scope
// see Example for details
type Subscribable[TEventData any] interface {
	On(eventName string, subscriber Subscriber[TEventData]) Unsubscribe
	Subscribe(subscriber Subscriber[TEventData]) Unsubscribe
	SubscribeTo(subscriber Subscriber[TEventData], eventName ...string) Unsubscribe
}

// Events represent an object which provides event subscribing and event triggering options
type Events[TEventData any] struct {
	mtx  sync.Mutex
	list map[string]*Event[TEventData]
	all  Event[*multiEventData[TEventData]]
}

type multiEventData[TEventData any] struct {
	eventName string
	eventData TEventData
}

// On subscribes to specific event
func (c *Events[TEventData]) On(eventName string, subscriber Subscriber[TEventData]) Unsubscribe {
	c.mtx.Lock()

	c.lazyInit()

	eventSubscribers := c.getSubscribers(eventName)

	c.mtx.Unlock()

	return eventSubscribers.On(subscriber)
}

// Subscribe subscribes MultiSubscriber to all events which will be triggered
func (c *Events[TEventData]) Subscribe(subscriber MultiSubscriber[TEventData]) Unsubscribe {
	c.mtx.Lock()

	c.lazyInit()

	unsubscribe := c.all.On(func(data *multiEventData[TEventData]) {
		subscriber(data.eventName, data.eventData)
	})

	c.mtx.Unlock()

	return unsubscribe
}

// SubscribeTo like On but subscribes MultiSubscriber to several events
// returns Unsubscribe which calling will unsubscribe from all events
// call without eventNames is simple to Subscribe
func (c *Events[TEventData]) SubscribeTo(subscriber MultiSubscriber[TEventData], eventNames ...string) Unsubscribe {
	if 0 == len(eventNames) {
		return c.Subscribe(subscriber)
	}

	c.mtx.Lock()

	c.lazyInit()

	unsubscribes := make([]Unsubscribe, len(eventNames))

	for i, eventName := range eventNames {
		func(eventName string) {
			unsubscribes[i] = c.getSubscribers(eventName).On(func(data TEventData) {
				subscriber(eventName, data)
			})
		}(eventName)
	}

	c.mtx.Unlock()

	var once sync.Once

	return func() {
		once.Do(func() {
			for _, unsubscribe := range unsubscribes {
				unsubscribe()
			}
		})
	}
}

// Trigger calls every Subscriber of event eventName with data arg
// calls every MultiSubscriber with eventName and data args
func (c *Events[TEventData]) Trigger(eventName string, data TEventData) {
	c.mtx.Lock()

	c.lazyInit()

	event, ok := c.list[eventName]

	c.mtx.Unlock()

	if ok {
		event.Trigger(data)
	}

	c.all.Trigger(&multiEventData[TEventData]{
		eventName: eventName,
		eventData: data,
	})
}

// TriggerAsync is like the Trigger but calls every Subscriber in separated goroutine
func (c *Events[TEventData]) TriggerAsync(eventName string, data TEventData) {
	c.mtx.Lock()

	c.lazyInit()

	eventSubscribers, ok := c.list[eventName]

	c.mtx.Unlock()

	wg := &sync.WaitGroup{}

	if ok {
		eventSubscribers.triggerAsync(wg, data)
	}

	c.all.triggerAsync(wg, &multiEventData[TEventData]{
		eventName: eventName,
		eventData: data,
	})

	wg.Wait()
}

func (c *Events[TEventData]) lazyInit() {
	if nil == c.list {
		c.list = make(map[string]*Event[TEventData])
	}
}

func (c *Events[TEventData]) getSubscribers(eventName string) *Event[TEventData] {
	eventSubscribers, ok := c.list[eventName]

	if !ok {
		eventSubscribers = &Event[TEventData]{}

		c.list[eventName] = eventSubscribers
	}

	return eventSubscribers
}
