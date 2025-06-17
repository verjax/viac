package setup

import (
	"errors"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Config struct {
	GitRepo   string   `yaml:"git_repo,omitempty"`
	FilePaths []string `yaml:"file_paths,omitempty"`
	Verbose   bool     `yaml:"verbose"`
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

	if cfg.GitRepo == "" && len(cfg.FilePaths) == 0 {
		return Config{}, errors.New("either git_repo or file_paths is required in configuration")
	}

	return cfg, err
}
