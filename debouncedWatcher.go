package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// debouncedWatcher is a file watcher which de-bounces events
// i.e. events are accumulated and only flushed after the debounceTime has elapsed
type debouncedWatcher struct {
	*sync.WaitGroup
	eventQueue   *eventQueue
	debounceTime time.Duration
	watcher      *fsnotify.Watcher
}

// newDebounceWatcher returns a new debounceWatcher instance with the given
// debounceTime or returns an error
func newDebouncedWatcher(debounceTime time.Duration) (*debouncedWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &debouncedWatcher{
		WaitGroup:    &sync.WaitGroup{},
		eventQueue:   newEventQueue(),
		debounceTime: debounceTime,
		watcher:      watcher,
	}, nil
}

// add adds a file to be watched
func (w *debouncedWatcher) add(file string) error {
	return w.watcher.Add(file)
}

// close closes the debouncedWatcher. This should be called whenever the
// debouncedWatcher is not used anymore
func (w *debouncedWatcher) close() error {
	return w.watcher.Close()
}

// watchAsync starts watching the registered files in a separate go-routine.
// It returns a cancel function which can be called to stop watching.
func (w *debouncedWatcher) watchAsync() func() error {
	ctx, cancel := context.WithCancel(context.Background())
	w.Add(1)
	go w.watch(ctx)
	return func() error {
		cancel()
		return w.close()
	}
}

// watch watches the registered files. It stops when ctx is done.
func (w *debouncedWatcher) watch(ctx context.Context) {
	events := newEventQueue()
	debounceTicker := time.NewTicker(time.Second * debounceTimeInSeconds)

	processElement := func(event fsnotify.Event) {
		fmt.Printf("DEBOUNCE: %v %v\n", event.Op, event.Name)
	}
	quit := func() {
		debounceTicker.Stop()
		events.flush(processElement)
		w.Done()
		fmt.Printf("DONE\n")
	}
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				quit()
				return
			}
			events.queue(event)
		case <-debounceTicker.C:
			events.flush(processElement)
		case err, ok := <-w.watcher.Errors:
			if !ok {
				quit()
				return
			}
			fmt.Printf("ERR: %v\n", err)
		case <-ctx.Done():
			quit()
			return
		}
	}
}
