package database

import (
	"os"
	"path"
	"testing"

	"github.com/p-program/kube-killer/config"
)

var projectConfig *config.ProjectConfig

// var mysqlConfig *config.MysqlConfig
var mySQL *MysqlPreparation

func TestMain(m *testing.M) {
	var err error
	projectConfig, err = prepareConfig()
	if err != nil {
		panic(err)
	}
	mysqlConfig := &projectConfig.Database.Mysql
	mySQL = NewMysqlPreparation(mysqlConfig)
	os.Exit(m.Run())
}

func prepareConfig() (*config.ProjectConfig, error) {
	config := config.NewProjectConfig()
	path := path.Join("../", "../", "../", "config", "config.yaml")
	err := config.LoadYAML(path)
	return config, err
}

func TestPrepare(t *testing.T) {
	mySQL.Prepare()
}

func TestCleanUp(t *testing.T) {
	mySQL.CleanUp()
}
