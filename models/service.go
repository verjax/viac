package models

type Service struct {
	Image     string   `yaml:"image"`
	DependsOn []string `yaml:"depends_on"`
	Ports     []string `yaml:"ports"`
}
