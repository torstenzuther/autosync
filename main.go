package main

import (
	"bufio"
	"log"
	"os"
	"time"
)

const (
	debounceTimeInSeconds = 10
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please add path as argument")
	}
	debouncedWatcher, err := newDebouncedWatcher(time.Second * debounceTimeInSeconds)
	if err != nil {
		log.Fatal(err)
	}
	defer debouncedWatcher.close()
	err = debouncedWatcher.add(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	cancel := debouncedWatcher.watchAsync()
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
