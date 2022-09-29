package main

import (
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
func newWatcherSwarm(watcherFactory func(string, ProcessingConfig) (watcher, error), store store) *watcherSwarm {
	return &watcherSwarm{
		watchers:       []watcher{},
		watcherFactory: watcherFactory,
		store:          store,
	}
}

// updateWatchers reconfigures the watcherSwarm by closing and recreating watchers
func (w *watcherSwarm) updateWatchers(config *Config) error {
	w.close()
	w.watchers = nil
	for _, pathMapping := range config.PathMappings {
		configPathPatternAbs, err := filepath.Abs(pathMapping.Pattern)
		if err != nil {
			return err
		}
		watchPath := filepath.Join(filepath.Dir(configPathPatternAbs), "...")
		watcher, err := w.watcherFactory(watchPath, config.Processing)
		if err != nil {
			return err
		}
		watcher.watchAsync(processFunc(w.store, pathMapping.GitPath, configPathPatternAbs), func() {
			if err := w.store.commit(); err != nil {
				logError(err)
			}
			if err := w.store.push(); err != nil {
				logError(err)
			}
		})
		w.watchers = append(w.watchers, watcher)
		log.Printf("Watching %v\n", configPathPatternAbs)
	}
	return nil
}

// close the watcherSwarm i.e. all its watchers
func (w *watcherSwarm) close() {
	for _, watcher := range w.watchers {
		watcher.close()
	}
}

func logError(err error) {
	log.Printf("%v\n", err)
}

func processFunc(store store, alias string, pattern string) func(string, notify.Event) {
	return func(eventPath string, event notify.Event) {
		patternAbs, err := filepath.Abs(pattern)
		if err != nil {
			logError(err)
			return
		}
		ok, err := filepath.Match(patternAbs, eventPath)
		if err != nil {
			logError(err)
			return
		}
		if ok {
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
				logError(err)
			}
		}
	}
}
