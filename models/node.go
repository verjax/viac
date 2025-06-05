package models

type Node struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
}
