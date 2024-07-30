package main

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config the configuration of the application
type Config struct {
	AlertmanagerURL      string   `yaml:"alertmanager_url"`
	AlertmanagerUsername string   `yaml:"alertmanager_username"`
	AlertmanagerPassword string   `yaml:"alertmanager_password"`
	ExpiredSilences      bool     `yaml:"expired_silences,omitempty"`
	Tenants              []string `yaml:"tenants,omitempty"`
}

func loadConfig(path string) (*Config, error) {
	conf := &Config{}
	if path != "" {
		f, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal([]byte(f), &conf)
		if err != nil {
			return nil, err
		}
	}

	envURL := os.Getenv("ALERTMANAGER_URL")
	if envURL != "" {
		conf.AlertmanagerURL = envURL
	}

	envUsername := os.Getenv("ALERTMANAGER_USERNAME")
	if envUsername != "" {
		conf.AlertmanagerUsername = envUsername
	}

	envPassword := os.Getenv("ALERTMANAGER_PASSWORD")
	if envPassword != "" {
		conf.AlertmanagerPassword = envPassword
	}

	log.Printf("Config loaded, path: %s", path)

	return conf, nil
}
