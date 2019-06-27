package app

import (
	"encoding/json"
	"os"
)

type Config struct {
	LibraryPath string        `json:"library"`
	Server      *ServerConfig `json:"server"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func DefaultConfig() *Config {
	return &Config{
		LibraryPath: "videos",
		Server: &ServerConfig{
			Host: "127.0.0.1",
			Port: 0,
		},
	}
}

func (c *Config) ReadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	return d.Decode(c)
}
