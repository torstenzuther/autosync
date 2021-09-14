package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

const (
	comment   = "#"
	separator = ":"
)

type config struct {
	// paths maps alias to actual path
	paths map[string]string
}

func parseConfig(reader io.Reader) (*config, error) {
	scanner := bufio.NewScanner(reader)
	result := &config{paths: map[string]string{}}
	for scanner.Scan() {
		trimmed := bytes.TrimSpace(scanner.Bytes())
		if bytes.HasPrefix(trimmed, []byte(comment)) {
			continue
		}
		splitted := bytes.Split(trimmed, []byte(separator))
		var alias string
		var file string
		if len(splitted) == 2 {
			alias = string(bytes.TrimSpace(splitted[0]))
			file = string(bytes.TrimSpace(splitted[1]))
		} else if len(splitted) == 1 {
			alias = string(bytes.TrimSpace(splitted[0]))
			file = alias
		} else if len(splitted) == 0 {
			continue
		} else {
			return nil, errors.New(fmt.Sprintf("Malformed line: %v", splitted))
		}
		if _, ok := result.paths[alias]; ok {
			return nil, errors.New(fmt.Sprintf("Alias already exists: %v", alias))
		}
		// TODO check parent paths (topological sort)
		// TODO glob patterns
		result.paths[alias] = file
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
