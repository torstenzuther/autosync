package main

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
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

	if err := store.commit(); err != nil {
		log.Fatal(err)
	}
	if err := store.push(); err != nil && err != git.NoErrAlreadyUpToDate {
		log.Fatal(err)
	}
	watcherSwarm := newWatcherSwarm(debouncedWatcherFactory, store)
	if err := watcherSwarm.updateWatchers(config); err != nil {
		log.Fatal(err)
	}
	defer watcherSwarm.close()

	if _, _, err := bufio.NewReader(os.Stdin).ReadRune(); err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
	}
	if err := store.push(); err != nil && err != git.NoErrAlreadyUpToDate {
		log.Fatal(err)
	}
}
