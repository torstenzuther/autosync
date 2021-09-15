package main

import (
	"fmt"
	"sync"

	"github.com/rjeczalik/notify"
)

// eventQueue holds events and embeds a sync.WaitGroup whose
// counter equals the number of elements in the queue
type eventQueue struct {
	*sync.WaitGroup
	events map[string]notify.Event
}

// newEventQueue creates an empty eventQueue
func newEventQueue() *eventQueue {
	return &eventQueue{
		WaitGroup: &sync.WaitGroup{},
		events:    map[string]notify.Event{},
	}
}

// flush empties the eventQueue and processes each element by
// passing it to the provided callback function
func (q *eventQueue) flush(processCallback func(path string, event notify.Event)) {
	for path, event := range q.events {
		processCallback(path, event)
	}
	q.Add(-len(q.events))
	q.events = map[string]notify.Event{}
}

// queue adds an event to the eventQueue
func (q *eventQueue) queue(event notify.EventInfo) {
	if _, ok := q.events[event.Path()]; !ok {
		q.Add(1)
	}
	q.events[event.Path()] = event.Event()
	fmt.Printf("EVENT: %v %v\n", event.Path(), q.events[event.Path()])
}
