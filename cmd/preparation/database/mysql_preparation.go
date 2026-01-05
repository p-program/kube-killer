package database

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/rs/zerolog/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/p-program/kube-killer/config"
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
	config *config.MysqlConfig
}

func NewMysqlPreparation(config *config.MysqlConfig) *MysqlPreparation {
	p := MysqlPreparation{
		config: config,
	}
	return &p
}

func (p *MysqlPreparation) Prepare() {
	db := p.getDBWithoutClose()
	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			defer sqlDB.Close()
		}
	}
	config := p.config
	err := p.CreateDb(db, config.Db)
	if err != nil {
		log.Err(err)
		return
	}
	err = p.CreateTable(db)
	if err != nil {
		log.Err(err)
		return
	}
}

// CleanUp delete database
func (p *MysqlPreparation) CleanUp() {
	dbName := p.config.Db
	sql := fmt.Sprintf("truncate table `%s`", dbName)
	log.Warn().Msgf("sql: %s", sql)
	db := p.getDBWithoutClose()
	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			defer sqlDB.Close()
		}
		db.Exec(sql)
	}
}

// CreateTable params are optional
func (p *MysqlPreparation) CreateTable(db *gorm.DB) error {
	// load template
	content, err := ioutil.ReadFile(MYSQL_TABLE_TERMINATOR_PATH)
	if err != nil {
		return err
	}
	dbName := p.config.Db
	sql := strings.ReplaceAll(string(content), "@db", dbName)
	tableName := p.config.Table
	sql = strings.ReplaceAll(string(content), "@table", tableName)
	log.Info().Msgf("sql: %s", sql)
	// db.Exec(sql)
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
	sql := strings.ReplaceAll(string(content), "@db", dbName)
	log.Info().Msgf("sql: %s", sql)
	db.Exec(sql)
	return nil
}

// func (m *MysqlPreparation) getDB() *gorm.DB {
// 	mysqlConfig := m.config
// 	mysqlConnectionString := fmt.Sprintf(MYSQL_TEMPLATE, mysqlConfig.User, mysqlConfig.Pwd, mysqlConfig.Host, mysqlConfig.Db)
// 	db, err := gorm.Open("mysql", mysqlConnectionString)
// 	// defer db.Close()
// 	if err != nil {
// 		log.Err(err)
// 		return nil
// 	}
// 	return db
// }

// getDBWithoutClose WARNING: please close db manually
func (m *MysqlPreparation) getDBWithoutClose() *gorm.DB {
	mysqlConfig := m.config
	mysqlConnectionString := fmt.Sprintf(MYSQL_TEMPLATE, mysqlConfig.User, mysqlConfig.Pwd, mysqlConfig.Host, mysqlConfig.Db)
	db, err := gorm.Open(mysql.Open(mysqlConnectionString), &gorm.Config{})
	if err != nil {
		log.Err(err)
		return nil
	}
	return db
}
