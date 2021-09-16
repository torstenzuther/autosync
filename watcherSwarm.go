package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/rjeczalik/notify"
)

type watcherSwarm struct {
	watchers       map[string]watcher
	watcherFactory watcherFactory
}

func newWatcherSwarm(watcherFactory func(string) (watcher, error)) *watcherSwarm {
	return &watcherSwarm{
		watchers:       map[string]watcher{},
		watcherFactory: watcherFactory,
	}
}

func (w *watcherSwarm) updateWatchers(config *config) {
	w.close()
	w.watchers = map[string]watcher{}
	for alias, configPathPattern := range config.paths {
		configPathPatternAbs, err := filepath.Abs(configPathPattern)
		if err != nil {
			log.Fatal(err)
		}
		watchPath := filepath.Join(filepath.Dir(configPathPatternAbs), "...")
		if _, ok := w.watchers[watchPath]; ok {
			w.watchers[watchPath].close()
		}
		watcher, err := w.watcherFactory(watchPath)
		if err != nil {
			log.Printf("%v\n", err)
			continue
		}
		watcher.watchAsync(processFunc(alias, configPathPatternAbs))
		w.watchers[watchPath] = watcher
		fmt.Printf("Watching %v\n", configPathPatternAbs)
	}
}

func (w *watcherSwarm) close() {
	for _, watcher := range w.watchers {
		watcher.close()
	}
}

func processFunc(alias string, pattern string) func(string, notify.Event) {
	return func(eventPath string, event notify.Event) {
		patternAbs, err := filepath.Abs(pattern)
		if err != nil {
			log.Printf("%v\n", err)
		}
		ok, err := filepath.Match(patternAbs, eventPath)
		if err != nil {
			log.Printf("%v\n", err)
		}
		if ok {
			fmt.Printf(": %v %v %v -> %v\n", eventPath, patternAbs, event, alias)
		}
	}
}
