package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	debounceTimeInSeconds = 10
)

func main() {
	configFile, err := os.Open("./config")
	if err != nil {
		log.Fatal(err)
	}
	config, err := parseConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	debouncedWatcher := newDebouncedWatcher(time.Second * debounceTimeInSeconds)
	for _, path := range config.paths {
		path, err := filepath.Abs(path)
		if err != nil {
			log.Fatal(err)
		}
		path = filepath.Join(filepath.Dir(path), "...")
		fmt.Printf("Adding %v\n", path)
		if err := debouncedWatcher.add(path); err != nil {
			log.Printf("%v\n", err)
		}
	}
	debouncedWatcher.watchAsync()
	defer debouncedWatcher.close()

	reader := bufio.NewReader(os.Stdin)
	_, _, err = reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}
}
