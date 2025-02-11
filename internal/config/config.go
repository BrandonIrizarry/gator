package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

// Read the current configuration into a new Config struct which is
// returned.
func Read() (config Config, err error) {
	filename, err := getConfigFilePath()

	if err != nil {
		return
	}

	file, err := os.Open(filename)
	defer file.Close()

	decoder := json.NewDecoder(file)

	if err = decoder.Decode(&config); err != nil {
		return
	}

	// 'config' should contain data, and err should be nil.
	return
}

// Set the username in the configuration.
func (config *Config) SetUser(username string) error {
	filename, err := getConfigFilePath()

	if err != nil {
		return err
	}

	config.CurrentUserName = username
	buffer := new(bytes.Buffer)

	encoder := json.NewEncoder(buffer)

	if err = encoder.Encode(config); err != nil {
		return err
	}

	if err = os.WriteFile(filename, buffer.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}

// Get the full path to the config file.
func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", home, configFileName), nil
}
