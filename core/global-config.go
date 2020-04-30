package core

import (
	"io/ioutil"

	"github.com/p-program/go-common-library/text"
	"gopkg.in/yaml.v2"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Name      string
	Namespace string
}

type DatabaseConfig struct {
	Mysql MysqlConfig
}

type MysqlConfig struct {
	Db    string
	Table string
	Host  string
	User  string
	Pwd   string
}

func NewConfig() *Config {
	return &Config{}
}

func (config *Config) LoadYAML(path string) error {
	content, err := ioutil.ReadFile(path)
	contentWithoutComment := []byte(text.RemoveYAMLcomment(string(content)))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(contentWithoutComment, &config)
	return err
}
