package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/rjeczalik/notify"
)

const eventChannelSize = 10000

// debouncedWatcher is a file watcher which de-bounces events
// i.e. events are accumulated and only flushed after the debounceTime has elapsed
type debouncedWatcher struct {
	*sync.WaitGroup
	eventQueue   *eventQueue
	events       chan notify.EventInfo
	debounceTime time.Duration
}

// newDebounceWatcher returns a new debouncedWatcher instance with the given
// debounceTime
func newDebouncedWatcher(debounceTime time.Duration) *debouncedWatcher {
	return &debouncedWatcher{
		WaitGroup:    &sync.WaitGroup{},
		eventQueue:   newEventQueue(),
		debounceTime: debounceTime,
		events:       make(chan notify.EventInfo, eventChannelSize),
	}
}

// add a file or folder to be watched
func (w *debouncedWatcher) add(file string) error {
	return notify.Watch(file, w.events, notify.All)
}

// close the debouncedWatcher. This should be called whenever the
// debouncedWatcher is not used anymore.
func (w *debouncedWatcher) close() {
	notify.Stop(w.events)
	close(w.events)
	w.Wait()
}

// watchAsync starts watching the registered files in a separate go-routine.
func (w *debouncedWatcher) watchAsync() {
	w.Add(1)
	go w.watch()
}

// watch the registered files.
func (w *debouncedWatcher) watch() {
	events := newEventQueue()
	debounceTicker := time.NewTicker(time.Second * debounceTimeInSeconds)
	processElement := func(path string, event notify.Event) {
		fmt.Printf("DEBOUNCE: %v %v\n", path, event)
	}
	quit := func() {
		debounceTicker.Stop()
		events.flush(processElement)
		w.Done()
		fmt.Printf("DONE\n")
	}
	for {
		select {
		case event, ok := <-w.events:
			if !ok {
				quit()
				return
			}
			events.queue(event)
		case <-debounceTicker.C:
			events.flush(processElement)
		}
	}
}
