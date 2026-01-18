package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error: %v\n", err)
	}
	configPath := filepath.Join(homeDir, configFileName)

	return configPath, nil
}

func Read() (Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return Config{}, fmt.Errorf("error: %v\n", err)
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("error: %v\n", err)
	}

	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, fmt.Errorf("error: %v\n", err)
	}

	// Version with decoder instead of Unmarshal
	// file, err := os.Open(configPath)
	// if err != nil {
	// 	return Config{}, err
	// }
	// defer file.Close()
	//
	// decoder := json.NewDecoder(file)
	// cfg := Config{}
	// err = decoder.Decode(&cfg)
	// if err != nil {
	// 	return Config{}, err
	// }

	return config, nil
}

func (cfg *Config) SetUser(userName string) error {
	if userName == "" {
		return fmt.Errorf("No username provided.\n")
	}

	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("error: %v\n", err)
	}

	cfg.CurrentUserName = userName

	file, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error: %v\n", err)
	}

	err = os.WriteFile(configPath, file, 0644)
	if err != nil {
		return fmt.Errorf("error: %v\n", err)
	}

	// Version with decoder instead of Marshal
	// file, err := os.Create(fullPath)
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()
	//
	// encoder := json.NewEncoder(file)
	// err = encoder.Encode(cfg)
	// if err != nil {
	// 	return err
	// }
	return nil
}
