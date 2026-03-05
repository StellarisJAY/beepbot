package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func FromConfigFile[T any](path string) (*T, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file error: %w", err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read config file error: %w", err)
	}

	config := new(T)
	if err := json.Unmarshal(bytes, config); err != nil {
		return nil, fmt.Errorf("parse config file error: %w", err)
	}
	return config, nil
}
