package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the configuration file structure
type FileConfig struct {
	File       string   `yaml:"file"`
	Required   []string `yaml:"required"`
	Example    string   `yaml:"example"`
	Strict     bool     `yaml:"strict"`
	CheckLeaks bool     `yaml:"check_leaks"`
	Quiet      bool     `yaml:"quiet"`
	JSON       bool     `yaml:"json"`
	GitHub     bool     `yaml:"github"`
	Ignore     []string `yaml:"ignore"`
	NoColor    bool     `yaml:"no_color"`
}

// configFileNames lists the supported config file names in priority order
var configFileNames = []string{
	".env-audit.yaml",
	".env-audit.yml",
}

// LoadFile loads configuration from a YAML file
func LoadFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// FindConfigFile looks for a config file in the current directory
// Returns the path if found, empty string if not found
func FindConfigFile() string {
	for _, name := range configFileNames {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}
	return ""
}

// FindConfigFileInDir looks for a config file in the specified directory
func FindConfigFileInDir(dir string) string {
	for _, name := range configFileNames {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

