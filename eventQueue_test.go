package main

import (
	"testing"

	"github.com/rjeczalik/notify"
	"github.com/stretchr/testify/assert"
)

type mockEvent struct {
	event notify.Event
	path  string
	sys   interface{}
}

func (m mockEvent) Event() notify.Event {
	return m.event
}

func (m mockEvent) Path() string {
	return m.path
}

func (m mockEvent) Sys() interface{} {
	return m.sys
}

func TestNewEventQueue(t *testing.T) {
	eventQueue := newEventQueue()
	assert.NotNil(t, eventQueue)
	assert.NotNil(t, eventQueue.events)
	assert.NotNil(t, eventQueue.WaitGroup)
	assert.Empty(t, eventQueue.events)
}

func TestQueueFlush(t *testing.T) {
	type testCase struct {
		expectedProcessedPaths  []string
		expectedProcessedEvents []notify.EventInfo
	}

	for _, test := range []testCase{
		{
			expectedProcessedPaths:  nil,
			expectedProcessedEvents: []notify.EventInfo{},
		},
		{
			expectedProcessedPaths: []string{"p1"},
			expectedProcessedEvents: []notify.EventInfo{
				mockEvent{
					event: 120,
					path:  "p1",
					sys:   "12",
				},
			},
		},
		{
			expectedProcessedPaths: []string{"p1", "p2"},
			expectedProcessedEvents: []notify.EventInfo{
				mockEvent{
					event: 120,
					path:  "p1",
					sys:   "12",
				},
				mockEvent{
					event: 1,
					path:  "p2",
				},
			},
		},
	} {
		var paths []string
		var events []notify.Event
		var expectedEvents []notify.Event
		var actualEvents []notify.Event

		eventQueue := newEventQueue()
		for _, event := range test.expectedProcessedEvents {
			eventQueue.queue(event)
			expectedEvents = append(expectedEvents, event.Event())
		}
		for _, event := range eventQueue.events {
			actualEvents = append(actualEvents, event)
		}
		assert.EqualValues(t, expectedEvents, actualEvents)
		eventQueue.flush(func(path string, event notify.Event) {
			paths = append(paths, path)
			events = append(events, event)
		})
		assert.Empty(t, eventQueue.events)
		assert.EqualValues(t, paths, test.expectedProcessedPaths)
		assert.EqualValues(t, events, expectedEvents)
	}
}
