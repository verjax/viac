package parser

import (
	"gopkg.in/yaml.v2"
	"os"
	"viac/models"
)

func ParseYAML(filepath string) (models.Node, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return models.Node{}, err
	}

	var compose models.Node
	err = yaml.Unmarshal(data, &compose)
	return compose, err
}
