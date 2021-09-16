package main

import (
	"bufio"
	"log"
	"os"
)

const (
	configPath = "./config"
)

func main() {
	config := mustReadConfig(configPath)
	watcherSwarm := newWatcherSwarm(debouncedWatcherFactory)
	watcherSwarm.updateWatchers(config)
	defer watcherSwarm.close()

	if _, _, err := bufio.NewReader(os.Stdin).ReadRune(); err != nil {
		log.Fatal(err)
	}
}
