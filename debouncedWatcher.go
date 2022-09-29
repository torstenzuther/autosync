package main

import (
	"sync"
	"time"

	"github.com/rjeczalik/notify"
)

// debouncedWatcher is a file watcher which de-bounces events
// i.e. events are accumulated and only flushed after the debounceTime has elapsed
type debouncedWatcher struct {
	*sync.WaitGroup
	events       chan notify.EventInfo
	debounceTime time.Duration
	config       ProcessingConfig
}

// newDebounceWatcher returns a new debouncedWatcher instance with the given
// file and debounceTime. It immediately starts watching or returns an error.
func newDebouncedWatcher(file string, debounceTime time.Duration, eventChannelSize int) (*debouncedWatcher, error) {
	events := make(chan notify.EventInfo, eventChannelSize)
	if err := notify.Watch(file, events, notify.All); err != nil {
		defer close(events)
		defer notify.Stop(events)
		return nil, err
	}

	return &debouncedWatcher{
		WaitGroup:    &sync.WaitGroup{},
		events:       events,
		debounceTime: debounceTime,
	}, nil
}

// close the debouncedWatcher. This should be called whenever the
// debouncedWatcher is not used anymore.
func (w *debouncedWatcher) close() {
	notify.Stop(w.events)
	close(w.events)
	w.Wait()
}

// watchAsync starts watching the registered files in a separate go-routine.
func (w *debouncedWatcher) watchAsync(processElement func(path string, event notify.Event), flush func()) {
	w.Wait()
	w.Add(1)
	go w.watch(processElement, flush)
}

// watch the registered files and call the given callback function for each event
func (w *debouncedWatcher) watch(processElement func(path string, event notify.Event), flush func()) {
	events := newEventQueue()
	debounceTicker := time.NewTicker(w.debounceTime)
	quit := func() {
		defer w.Done()
		debounceTicker.Stop()
		events.flush(processElement, flush)
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
			events.flush(processElement, flush)
		}
	}
}

func debouncedWatcherFactory(watchPath string, config ProcessingConfig) (watcher, error) {
	duration, err := time.ParseDuration(config.DebounceDuration)
	if err != nil {
		return nil, err
	}
	return newDebouncedWatcher(watchPath, duration, config.EventChannelSize)
}
