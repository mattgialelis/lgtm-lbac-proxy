package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	LokiURL       string `yaml:"lokiUrl"`
	logType       string `yaml:"logType"`
	LokiBasicAuth struct {
		Enabled  bool   `yaml:"enabled"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"lokiBasicAuth"`
	AdminUser struct {
		Username string `yaml:"username"`
	} `yaml:"adminUser"`
}

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return nil, fmt.Errorf("CONFIG_PATH environment variable is not set")
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
