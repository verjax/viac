package setup

import (
	"errors"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Config struct {
	FilePaths   []string `yaml:"file_paths"`
	GitRepos    []string `yaml:"git_repos"`
	Directories []string `yaml:"directories"`
	AutoDetect  bool     `yaml:"auto_detect"`
	Verbose     bool     `yaml:"verbose"`
}

func LoadConfig(path string) (Config, error) {
	if ext := filepath.Ext(path); ext != ".yaml" && ext != ".yml" {
		return Config{}, errors.New("unsupported config file format: only .yaml or .yml is allowed")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}
