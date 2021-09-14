package main

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// eventQueue holds events and embeds a sync.WaitGroup which is
// holds the number of elements in the queue
type eventQueue struct {
	*sync.WaitGroup
	events map[string]fsnotify.Event
}

// newEventQueue creates an empty eventQueue
func newEventQueue() *eventQueue {
	return &eventQueue{
		WaitGroup: &sync.WaitGroup{},
		events:    map[string]fsnotify.Event{},
	}
}

// flush empties the eventQueue and processes each element by
// passing it to the provided callback function
func (q *eventQueue) flush(processCallback func(event fsnotify.Event)) {
	for _, event := range q.events {
		processCallback(event)
	}
	q.Add(-len(q.events))
	q.events = map[string]fsnotify.Event{}
}

// queue adds an event to the eventQueue
func (q *eventQueue) queue(event fsnotify.Event) {
	if _, ok := q.events[event.Name]; !ok {
		q.Add(1)
	}
	q.events[event.Name] = event
	fmt.Printf("EVENT: %#v %v\n", q.events[event.Name], event.Op)
}
