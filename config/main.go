package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	ConfigPath string
)

type Config struct {
	NumReplicas       int               `json:"num_replicas"`
	Byzantine         bool              `json:"byzantine"`
	APIServerAddr     string            `json:"server_addr"`
	LogConfig         LogConfig         `json:"log"`
	ReportStoreConfig ReportStoreConfig `json:"report_store"`
}

type LogConfig struct {
	Path   string `json:"path"`
	Format string `json:"format"`
	// one of panic|fatal|error|warn|warning|info|debug|trace
	Level string `json:"level"`
}

type ReportStoreConfig struct {
	Path string `json:"path"`
}

func ParseConfig(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}
	var defaultConfig = &Config{
		NumReplicas:   4,
		Byzantine:     true,
		APIServerAddr: "0.0.0.0:7074",
		LogConfig: LogConfig{
			Path:   "",
			Format: "json",
			Level:  "info",
		},
		ReportStoreConfig: ReportStoreConfig{
			Path: "test_reports",
		},
	}
	err = json.Unmarshal(bytes, &defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %s", err)
	}
	return defaultConfig, nil
}
