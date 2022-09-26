package main

import "github.com/rjeczalik/notify"

type watcher interface {
	close()
	watchAsync(processElement func(path string, event notify.Event), flush func())
}

type watcherFactory func(watchPath string) (watcher, error)
