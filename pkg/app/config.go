package app

import (
	"encoding/json"
	"os"
)

type Config struct {
	LibraryPath string        `json:"library"`
	Server      *ServerConfig `json:"server"`
	Feed        *FeedConfig   `json:"feed"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type FeedConfig struct {
	ExternalURL string `json:"external_url"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Author      struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
	Copyright string `json:"copyright"`
}

func DefaultConfig() *Config {
	return &Config{
		LibraryPath: "videos",
		Server: &ServerConfig{
			Host: "127.0.0.1",
			Port: 0,
		},
		Feed: &FeedConfig{
			ExternalURL: "http://localhost",
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
