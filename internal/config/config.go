package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DB_Url            string `json:"db_url"`
	Current_User_Name string `json:"current_user_name"`
}

const filename = "/.gatorconfig.json"

func Read() (Config, error) {
	fullPath, err := getConfigFilePath()
	file, err := os.Open(fullPath)
	if err != nil {
		return Config{}, fmt.Errorf("Couldn't open the file: %s", fullPath)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var cfg Config
	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't decode the config json")
	}

	return cfg, nil
}

func (c *Config) SetUser(username string) error {
	c.Current_User_Name = username
	return write(*c)
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir + filename, nil
}

func write(cfg Config) error {
	fullPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}
