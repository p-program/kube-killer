package database

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/p-program/kube-killer/core"
)

const (
	MYSQL_PATH                  = "/deploy/database/mysql/"
	MYSQL_DB_Path               = MYSQL_PATH + "db_kube_killer.sql"
	MYSQL_TABLE_TERMINATOR_PATH = MYSQL_PATH + "table_terminator.sql"
	DEFAULT_DB_NAME             = "kube_killer"
	DEFAULT_TABLE_NAME          = "terminator"
	// MYSQL_TEMPLATE example : "user:password@tcp(192.168.1.4:3306)/dbname?charset=utf8&parseTime=True&loc=Local"
	MYSQL_TEMPLATE = "%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local"
)

type MysqlPreparation struct {
	config *core.MysqlConfig
}

func NewMysqlPreparation(config *core.MysqlConfig) *MysqlPreparation {
	p := MysqlPreparation{
		config: config,
	}
	return &p
}

func (p *MysqlPreparation) Prepare() {
	db := p.GetDBWithoutClose()
	defer db.Close()
	config := p.config
	err := p.CreateDb(db, config.Db)
	if err != nil {
		log.Err(err)
		return
	}
	err = p.CreateTable(db, config.Db, config.Table)
	if err != nil {
		log.Err(err)
		return
	}
}

func (p *MysqlPreparation) CleanUp(dbName string) {
	if dbName == "" {
		dbName = DEFAULT_DB_NAME
	}
	sql := fmt.Sprintf("truncate table `%s`" + dbName)
	db := p.GetDB()
	db.Exec(sql)
}

// CreateTable params are optional
func (p *MysqlPreparation) CreateTable(db *gorm.DB, dbName string, tableName string) error {
	content, err := ioutil.ReadFile(MYSQL_TABLE_TERMINATOR_PATH)
	if err != nil {
		return err
	}
	if dbName == "" {
		dbName = DEFAULT_DB_NAME
	}
	if tableName == "" {
		tableName = DEFAULT_TABLE_NAME
	}
	// db := p.GetDB()
	sql := strings.ReplaceAll(string(content), "@db", dbName)
	db.Exec(sql)
	return nil
}

// CreateDb params are optional
func (p *MysqlPreparation) CreateDb(db *gorm.DB, dbName string) error {
	content, err := ioutil.ReadFile(MYSQL_DB_Path)
	if err != nil {
		return err
	}
	if dbName == "" {
		dbName = DEFAULT_DB_NAME
	}
	// db := p.GetDB()
	sql := strings.ReplaceAll(string(content), "@db", dbName)
	db.Exec(sql)
	return nil
}

func (m *MysqlPreparation) GetDB() *gorm.DB {
	mysqlConfig := m.config
	mysqlConnectionString := fmt.Sprintf(MYSQL_TEMPLATE, mysqlConfig.User, mysqlConfig.Pwd, mysqlConfig.Host, mysqlConfig.Db)
	db, err := gorm.Open("mysql", mysqlConnectionString)
	defer db.Close()
	if err != nil {
		log.Err(err)
		return nil
	}
	return db
}

// GetDBWithoutClose WARNING: please close db manually
func (m *MysqlPreparation) GetDBWithoutClose() *gorm.DB {
	mysqlConfig := m.config
	mysqlConnectionString := fmt.Sprintf(MYSQL_TEMPLATE, mysqlConfig.User, mysqlConfig.Pwd, mysqlConfig.Host, mysqlConfig.Db)
	db, err := gorm.Open("mysql", mysqlConnectionString)
	if err != nil {
		log.Err(err)
		return nil
	}
	return db
}
