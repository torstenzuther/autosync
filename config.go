package main

import (
	"encoding/json"
	"os"
)

// GitConfig is the configuration for the git repository
type GitConfig struct {
	// Url is the url of the git repository
	Url string `json:"url"`
	// Auth is optional data for (HTTP Basic) authentication with the git repo
	Auth struct {
		// UserName is the username for the HTTP Basic Authentication
		UserName string `json:"username"`
		// Password is the password for the HTTP Basic Authentication
		Password string `json:"password"`
	} `json:"auth"`
}

// Config is the data structure for configuring Autosync
type Config struct {
	// GitRepo contains all git repository related configuration settings
	GitRepo GitConfig `json:"git-repo"`
	// PathMappings contains the mappings from git path to file pattern that is watched
	PathMappings []struct {
		// GitPath is the path (directory) of the git repository to put the files to
		GitPath string `json:"path"`
		// Pattern is the pattern that is watched. All file changes will be committed to the GitPath
		Pattern string `json:"pattern"`
	} `json:"path-mappings"`
}

// loadConfig loads the configuration from the given path (JSON format) or otherwise returns an error
func loadConfig(file string) (config *Config, err error) {
	var configFile *os.File
	configFile, err = os.Open(file)
	if err != nil {
		return
	}
	defer func(configFile *os.File) {
		err = configFile.Close()
	}(configFile)
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return
}
