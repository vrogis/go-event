package event_test

import (
	"fmt"
	event "go-event"
)

func ExampleEvents_On() {
	var events event.Events[int]

	unsubscribe := events.On("example-event", func(data int) {
		fmt.Println(data)
	})

	events.Trigger("example-event", 1)
	events.Trigger("example-event", 2)
	events.Trigger("example-event1", 3)

	unsubscribe()

	events.Trigger("example-event", 4)

	// Output: 1
	// 2
}

func ExampleEvents_Subscribe() {
	var events event.Events[int]

	unsubscribe1 := events.Subscribe(func(eventName string, data int) {
		fmt.Printf("1: %s %d\n", eventName, data)
	})

	unsubscribe2 := events.Subscribe(func(eventName string, data int) {
		fmt.Printf("2: %s %d\n", eventName, data)
	})

	events.Trigger("event", 1)
	events.Trigger("event2", 2)

	unsubscribe1()

	events.Trigger("event", 3)

	unsubscribe2()

	events.Trigger("event", 4)

	// Output:
	// 1: event 1
	// 2: event 1
	// 1: event2 2
	// 2: event2 2
	// 2: event 3
}

func ExampleEvents_SubscribeTo() {
	var events event.Events[int]

	unsubscribe1 := events.SubscribeTo(func(eventName string, data int) {
		fmt.Printf("1: %s %d\n", eventName, data)
	}, "event1", "event2")

	unsubscribe2 := events.SubscribeTo(func(eventName string, data int) {
		fmt.Printf("2: %s %d\n", eventName, data)
	}, "event2")

	events.Trigger("event1", 1)
	events.Trigger("event2", 2)

	unsubscribe1()

	events.Trigger("event1", 3)
	events.Trigger("event2", 4)

	unsubscribe2()

	events.Trigger("event1", 5)
	events.Trigger("event2", 6)

	// Output:
	// 1: event1 1
	// 1: event2 2
	// 2: event2 2
	// 2: event2 4
}

func ExampleEvents_TriggerAsync() {
	var events event.Events[int]

	unsubscribe1 := events.SubscribeTo(func(eventName string, data int) {
		fmt.Printf("1: %s %d\n", eventName, data)
	}, "event1", "event2")

	unsubscribe2 := events.SubscribeTo(func(eventName string, data int) {
		fmt.Printf("2: %s %d\n", eventName, data)
	}, "event2")

	events.TriggerAsync("event1", 1)
	events.TriggerAsync("event2", 2)

	unsubscribe1()

	events.TriggerAsync("event1", 3)
	events.TriggerAsync("event2", 4)

	unsubscribe2()

	events.TriggerAsync("event1", 5)
	events.TriggerAsync("event2", 6)

	// Unordered Output:
	// 1: event1 1
	// 1: event2 2
	// 2: event2 2
	// 2: event2 4
}
