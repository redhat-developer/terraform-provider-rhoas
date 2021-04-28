package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// NewFile creates a new config type
func NewFile() IConfig {
	cfg := &File{}

	return cfg
}

// File is a type which describes a config file
type File struct{}

// Load loads the configuration from the configuration file. If the configuration file doesn't exist
// it will return an empty configuration object.
func (c *File) Load() (*Config, error) {
	file, err := c.Location()
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	// #nosec G304
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save saves the given configuration to the configuration file.
func (c *File) Save(cfg *Config) error {
	file, err := c.Location()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, data, 0600)
	if err != nil {
		return err
	}
	return nil
}

// Remove removes the configuration file.
func (c *File) Remove() error {
	file, err := c.Location()
	if err != nil {
		return err
	}
	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		return nil
	}
	err = os.Remove(file)
	if err != nil {
		return err
	}
	return nil
}

// Location gets the path to the config file
func (c *File) Location() (path string, err error) {
	if rhoasConfig := os.Getenv("RHOASTF_CONFIG"); rhoasConfig != "" {
		path = rhoasConfig
	} else {
		path, err = getUserConfig(".rhoastf.json")
		if err != nil {
			return "", err
		}
	}
	return path, nil
}
