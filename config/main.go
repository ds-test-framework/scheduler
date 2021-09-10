package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	// ConfigPath is the variable which stores the config path command line parameter
	ConfigPath string
)

// Config stores the config for the tool
type Config struct {
	// NumReplicas number of replicas that are in the distributed system
	NumReplicas int `json:"num_replicas"`
	// Byzantine indicating if the algorithm being tested is byzantine fault tolerant
	Byzantine bool `json:"byzantine"`
	// APIServerAddr address of the APIServer
	APIServerAddr string `json:"server_addr"`
	// LogConfig configuration for logging
	LogConfig LogConfig `json:"log"`
	// ReportStoreConfig config for recording unit test reports
	ReportStoreConfig ReportStoreConfig `json:"report_store"`
}

// LogConfig stores the config for logging purpose
type LogConfig struct {
	// Path of the log file
	Path string `json:"path"`
	// Format to log. Only `json` is currently supported
	Format string `json:"format"`
	// Level log level, one of panic|fatal|error|warn|warning|info|debug|trace
	Level string `json:"level"`
}

// ReportStoreConfig configuration of the unit test reports
type ReportStoreConfig struct {
	Path    string `json:"path"`
	OldPath string `json:"old_path"`
}

// PatseConfig parses config from the specificied file
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
