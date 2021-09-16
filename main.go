package main

import (
	"bufio"
	"log"
	"os"
)

const (
	debounceTimeInSeconds = 1
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
	watcherSwarm := newWatcherSwarm(debouncedWatcherFactory)
	watcherSwarm.updateWatchers(config)
	defer watcherSwarm.close()

	reader := bufio.NewReader(os.Stdin)
	_, _, err = reader.ReadRune()
	if err != nil {
		log.Fatal(err)
	}
}
