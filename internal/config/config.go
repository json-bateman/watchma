package config

import (
	"encoding/json"
	"os"
)

type JellyfinConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

type AppConfig struct {
	Jellyfin JellyfinConfig `json:"jellyfin"`
	Server   struct {
		Port int `json:"port"`
	} `json:"server"`
}

var Config AppConfig

func Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(&Config)
}

func (a *AppConfig) ReturnApiKey() string {
	return a.Jellyfin.APIKey
}
