package database

import (
	"path"
	"testing"

	"github.com/p-program/kube-killer/core"
)

func prepareConfing(t *testing.T) *core.MysqlConfig {
	config := model.NewProjectConfig()
	path := path.Join("../", "config", "config.yaml")
	err := config.LoadYAML(path)
	assert.Nil(t, err)
	return config
}

func prepareMysqlConfig() (config *core.MysqlConfig) {

	return
}

func TestPrepare(t *testing.T) {
	config := prepareMysqlConfig()
	preparation := NewMysqlPreparation(config)
	preparation.Prepare()
}
