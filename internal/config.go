package internal

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Pipeline struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name    string            `yaml:"name"`
	Uses    string            `yaml:"uses,omitempty"`
	URL     string            `yaml:"url,omitempty"`
	Files   []string          `yaml:"files,omitempty"`
	Command string            `yaml:"command,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
}

type Config struct {
	Pipeline Pipeline `yaml:"pipeline"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
