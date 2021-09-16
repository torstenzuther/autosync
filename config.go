package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"os"
)

const (
	comment   = "#"
	separator = ":"
)

type config struct {
	// paths maps alias to actual path
	paths map[string]string
}

// parseConfig reads the config or returns an error
func parseConfig(reader io.Reader) (*config, error) {
	scanner := bufio.NewScanner(reader)
	result := &config{paths: map[string]string{}}
	for scanner.Scan() {
		trimmed := bytes.TrimSpace(scanner.Bytes())
		if len(trimmed) == 0 || bytes.HasPrefix(trimmed, []byte(comment)) {
			continue
		}
		split := bytes.Split(trimmed, []byte(separator))
		var alias string
		var file string
		switch len(split) {
		case 1:
			alias = string(bytes.TrimSpace(split[0]))
			file = alias
		case 2:
			alias = string(bytes.TrimSpace(split[0]))
			file = string(bytes.TrimSpace(split[1]))
		default:
			return nil, errors.New("malformed line")
		}
		if _, ok := result.paths[alias]; ok {
			return nil, errors.New("duplicated alias")
		}
		result.paths[alias] = file
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// mustReadConfig reads the config from the file system
// If there is an error it will panic
func mustReadConfig(path string) *config {
	configFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	config, err := parseConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	return config
}
