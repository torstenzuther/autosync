package main

import (
	"encoding/json"
	"os"
)

type GitConfig struct {
	Url  string `json:"url"`
	Auth struct {
		UserName string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

type Config struct {
	GitRepo      GitConfig `json:"git-repo"`
	PathMappings []struct {
		GitPath string `json:"path"`
		Pattern string `json:"pattern"`
	} `json:"path-mappings"`
}

func loadConfig(file string) (*Config, error) {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
