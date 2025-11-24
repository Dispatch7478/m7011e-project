package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Services []Service `yaml:"services"`
}

type Service struct {
	Name  string `yaml:"name"`
	URL   string `yaml:"url"`
	Proxy Proxy  `yaml:"proxy"`
}

type Proxy struct {
	Prefix  string `yaml:"prefix"`
	Rewrite string `yaml:"rewrite"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
