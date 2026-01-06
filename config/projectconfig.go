package config

import (
	"os"

	"github.com/p-program/go-common-library/text"
	"gopkg.in/yaml.v2"
)

type ProjectConfig struct {
	LogLevel string         `yaml:"logLevel"`
	Sever    ServerConfig   `yaml:"sever"`
	Database DatabaseConfig `yaml:"database"`
	// for copy
	// `yaml:""`
}

func NewProjectConfig() *ProjectConfig {
	return &ProjectConfig{}
}

// LoadYAML LoadYAML from file
func (config *ProjectConfig) LoadYAML(path string) error {
	content, err := os.ReadFile(path)
	contentWithoutComment := []byte(text.RemoveYAMLcomment(string(content)))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(contentWithoutComment, &config)
	if err != nil {
		return err
	}
	return err
}
