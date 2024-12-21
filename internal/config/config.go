package config

import (
	"encoding/json"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (*Config, error) {
	homedir, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	filename := homedir + configFileName
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) SetUser(usr string) {
	c.CurrentUserName = usr
	write(c)
}

func write(c *Config) error {
	homedir, err := getConfigPath()
	if err != nil {
		return err
	}

	jsondata, err := json.Marshal(c)
	if err != nil {
		return err
	}

	filename := homedir + configFileName

	os.WriteFile(filename, jsondata, 7777)
	return nil
}

func getConfigPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	retdir := homedir + "/"
	return retdir, nil
}
