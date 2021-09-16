package main

import (
	"sync"
	"time"

	"github.com/rjeczalik/notify"
)

const (
	eventChannelSize      = 10000
	debounceTimeInSeconds = 1
)

// debouncedWatcher is a file watcher which de-bounces events
// i.e. events are accumulated and only flushed after the debounceTime has elapsed
type debouncedWatcher struct {
	*sync.WaitGroup
	eventQueue   *eventQueue
	events       chan notify.EventInfo
	debounceTime time.Duration
}

// newDebounceWatcher returns a new debouncedWatcher instance with the given
// file and debounceTime. It immediately starts watching or returns an error.
func newDebouncedWatcher(file string, debounceTime time.Duration) (*debouncedWatcher, error) {
	events := make(chan notify.EventInfo, eventChannelSize)
	err := notify.Watch(file, events, notify.All)
	if err != nil {
		defer close(events)
		defer notify.Stop(events)
		return nil, err
	}
	w := &debouncedWatcher{
		WaitGroup:    &sync.WaitGroup{},
		eventQueue:   newEventQueue(),
		debounceTime: debounceTime,
		events:       events,
	}
	return w, nil
}

// close the debouncedWatcher. This should be called whenever the
// debouncedWatcher is not used anymore.
func (w *debouncedWatcher) close() {
	notify.Stop(w.events)
	close(w.events)
	w.Wait()
}

// watchAsync starts watching the registered files in a separate go-routine.
func (w *debouncedWatcher) watchAsync(processElement func(path string, event notify.Event)) {
	w.Add(1)
	go w.watch(processElement)
}

// watch the registered files and call the given callback function for each event
func (w *debouncedWatcher) watch(processElement func(path string, event notify.Event)) {
	events := newEventQueue()
	debounceTicker := time.NewTicker(time.Second * debounceTimeInSeconds)
	quit := func() {
		defer w.Done()
		debounceTicker.Stop()
		events.flush(processElement)
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

func debouncedWatcherFactory(watchPath string) (watcher, error) {
	return newDebouncedWatcher(watchPath, time.Second*debounceTimeInSeconds)
}
