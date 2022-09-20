package main

import (
	"bufio"
	"io"
	"log"
	"os"
)

const (
	configPath = "./config.json"
)

func main() {
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	store, err := newInMemoryStore(config)
	if err != nil {
		log.Fatal(err)
	}
	watcherSwarm := newWatcherSwarm(debouncedWatcherFactory, store)
	watcherSwarm.updateWatchers(config)
	defer watcherSwarm.close()

	if _, _, err := bufio.NewReader(os.Stdin).ReadRune(); err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
	}
	if err := store.push(); err != nil {
		log.Fatal(err)
	}
}
