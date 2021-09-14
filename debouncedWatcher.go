package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type debouncedWatcher struct {
	*sync.WaitGroup
	eventQueue   *eventQueue
	debounceTime time.Duration
}

func newDebouncedWatcher(debounceTime time.Duration) *debouncedWatcher {
	return &debouncedWatcher{
		WaitGroup:    &sync.WaitGroup{},
		eventQueue:   newEventQueue(),
		debounceTime: debounceTime,
	}
}

func (w *debouncedWatcher) watchAsync(watcher *fsnotify.Watcher) func() error {
	ctx, cancel := context.WithCancel(context.Background())
	w.Add(1)
	go w.watch(ctx, watcher)
	return func() error {
		cancel()
		return watcher.Close()
	}
}

func (w *debouncedWatcher) watch(ctx context.Context, watcher *fsnotify.Watcher) {
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
		case event, ok := <-watcher.Events:
			if !ok {
				quit()
				return
			}
			events.queue(event)
		case <-debounceTicker.C:
			events.flush(processElement)
		case err, ok := <-watcher.Errors:
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
