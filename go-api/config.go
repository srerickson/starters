package goapi

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	envVarSecret      = "API_SECRET"
	envVarDatabaseURL = "API_DB"
)

// Config holds main config options for the server
type Config struct {
	Secret       string
	DatabaseURL  string `yaml:"database"`
	ServerNet    string `yaml:"net"`
	ServerPort   int    `yaml:"port"`
	ServeTLS     bool   `yaml:"tls"`
	TLSCrt       string `yaml:"tls-crt"`
	TLSKey       string `yaml:"tls-key"`
	Prefix       string `yaml:"api-path-prefix"`
	StaticPrefix string `yaml:"static-path-prefix"`
	StaticDir    string `yaml:"static-dir"`
	Auths        []struct {
		ID        string `yaml:"id"`
		KeyDigest string `yaml:"key-digest"`
		Roles     []string
	} `yaml:"auths"`
	Verbose bool
}

// defaults
var config = Config{
	ServerNet:    "127.0.0.1",
	ServerPort:   8080,
	DatabaseURL:  "postgres://localhost",
	TLSCrt:       "server.crt",
	TLSKey:       "server.key",
	Prefix:       "/api",
	StaticDir:    "static",
	StaticPrefix: "/static",
}

// GetConfig returns a copy of the current config
func GetConfig() Config {
	return config
}

// LoadConfig returns
func LoadConfig(path string) (Config, error) {
	cfgRaw, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	if err = yaml.Unmarshal(cfgRaw, &config); err != nil {
		return config, err
	}
	// check environment for Secret and Database settings
	if ev := os.Getenv(envVarSecret); ev != "" {
		config.Secret = ev
	}
	if ev := os.Getenv(envVarDatabaseURL); ev != "" {
		config.DatabaseURL = ev
	}
	// check required settings
	if config.Secret == "" {
		return config, fmt.Errorf("%s not set", envVarSecret)
	}
	if config.DatabaseURL == "" {
		return config, fmt.Errorf("%s not set", envVarDatabaseURL)
	}
	if config.ServeTLS {
		config.ServerPort = 443
	}
	return config, nil
}
