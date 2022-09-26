package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/rjeczalik/notify"
)

// watcherSwarm is a set of watchers
type watcherSwarm struct {
	watchers       []watcher
	watcherFactory watcherFactory
	store          store
}

// newWatcherSwarm returns a new watcherSwarm instance
func newWatcherSwarm(watcherFactory func(string) (watcher, error), store store) *watcherSwarm {
	return &watcherSwarm{
		watchers:       []watcher{},
		watcherFactory: watcherFactory,
		store:          store,
	}
}

// updateWatchers reconfigures the watcherSwarm by closing and recreating watchers
func (w *watcherSwarm) updateWatchers(config *Config) {
	w.close()
	for _, watcher := range w.watchers {
		watcher.close()
	}
	w.watchers = nil
	for _, pathMapping := range config.PathMappings {
		configPathPatternAbs, err := filepath.Abs(pathMapping.Pattern)
		if err != nil {
			log.Fatal(err)
		}
		watchPath := filepath.Join(filepath.Dir(configPathPatternAbs), "...")
		watcher, err := w.watcherFactory(watchPath)
		if err != nil {
			log.Printf("%v\n", err)
			continue
		}
		watcher.watchAsync(processFunc(w.store, pathMapping.GitPath, configPathPatternAbs))
		w.watchers = append(w.watchers, watcher)
		fmt.Printf("Watching %v\n", configPathPatternAbs)
	}
}

// close the watcherSwarm i.e. all its watchers
func (w *watcherSwarm) close() {
	for _, watcher := range w.watchers {
		watcher.close()
	}
}

func processFunc(store store, alias string, pattern string) func(string, notify.Event) {
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

			var action func(string, string) error

			switch event {
			case notify.Create:
				action = store.onWrite
			case notify.Write:
				action = store.onWrite
			case notify.Rename:
				action = store.onRename
			case notify.Remove:
				action = store.onRemove
			default:
				return
			}

			if err := action(eventPath, alias); err != nil {
				log.Printf("%v\n", err)
			}
		}
	}
}
