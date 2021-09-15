package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rjeczalik/notify"
)

const (
	debounceTimeInSeconds = 1
)

type watcherSwarm struct {
	watchers map[string]*debouncedWatcher
}

func newWatcherSwarm() *watcherSwarm {
	return &watcherSwarm{watchers: map[string]*debouncedWatcher{}}
}

func main() {
	configFile, err := os.Open("./config")
	if err != nil {
		log.Fatal(err)
	}
	config, err := parseConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	watcherSwarm := newWatcherSwarm()
	watcherSwarm.updateWatchers(config)
	defer watcherSwarm.close()

	reader := bufio.NewReader(os.Stdin)
	_, _, err = reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}
}

func (w *watcherSwarm) updateWatchers(config *config) {
	w.close()
	w.watchers = map[string]*debouncedWatcher{}
	for alias, configPathPattern := range config.paths {
		configPathPatternAbs, err := filepath.Abs(configPathPattern)
		if err != nil {
			log.Fatal(err)
		}
		watchPath := filepath.Join(filepath.Dir(configPathPatternAbs), "...")
		if _, ok := w.watchers[watchPath]; ok {
			w.watchers[watchPath].close()
		}
		watcher, err := newDebouncedWatcher(watchPath, time.Second*debounceTimeInSeconds)
		if err != nil {
			log.Printf("%v\n", err)
			watcher.close()
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
