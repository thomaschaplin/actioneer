package parser

import (
	"os"

	"gopkg.in/yaml.v2"
	"thomaschaplin/github-actions-poc/types"
)

func ParseWorkflow(file string) (*types.Workflow, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var workflow types.Workflow
	err = yaml.Unmarshal(data, &workflow)
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}
