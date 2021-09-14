package main

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	debounceTimeInSeconds = 10
)

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
	_, _, err = reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}
	err = cancel()
	if err != nil {
		log.Fatal(err)
	}
	debouncedWatcher.Wait()
}
