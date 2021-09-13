package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	debounceTimeInSeconds = 10
)

type eventQueue struct {
	*sync.WaitGroup
	events map[string]fsnotify.Event
}

type debouncedWatcher struct {
	*sync.WaitGroup
	eventQueue   *eventQueue
	debounceTime time.Duration
}

func newEventQueue() *eventQueue {
	return &eventQueue{
		WaitGroup: &sync.WaitGroup{},
		events:    map[string]fsnotify.Event{},
	}
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

func (q *eventQueue) flush() {
	for fileName, event := range q.events {
		fmt.Printf("DEBOUNCE: %v %v\n", event.Op, fileName)
	}
	q.Add(-len(q.events))
	q.events = map[string]fsnotify.Event{}
}

func (q *eventQueue) queue(event fsnotify.Event) {
	if _, ok := q.events[event.Name]; !ok {
		q.Add(1)
	}
	q.events[event.Name] = event
	fmt.Printf("EVENT: %#v %v\n", q.events[event.Name], event.Op)
}

func (w *debouncedWatcher) watch(ctx context.Context, watcher *fsnotify.Watcher) {
	events := newEventQueue()
	debounceTicker := time.NewTicker(time.Second * debounceTimeInSeconds)

	quit := func() {
		debounceTicker.Stop()
		events.flush()
		w.Done()
		fmt.Printf("DONE\n")
	}
	for {
		select {
		case event := <-watcher.Events:
			events.queue(event)
		case <-debounceTicker.C:
			events.flush()
		case err := <-watcher.Errors:
			fmt.Printf("ERR: %v\n", err)
		case <-ctx.Done():
			quit()
			return
		}
	}
}

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) < 2 {
		log.Fatal("Please add path as argument")
	}
	err = watcher.Add(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	debouncedWatcher := newDebouncedWatcher(time.Second * debounceTimeInSeconds)
	cancel := debouncedWatcher.watchAsync(watcher)
	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()
	err = cancel()
	if err != nil {
		fmt.Errorf("%v\n", err)
	}
	debouncedWatcher.Wait()
}
