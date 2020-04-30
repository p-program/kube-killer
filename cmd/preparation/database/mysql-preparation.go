package database

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/p-program/kube-killer/core"
)

type MysqlPreparation struct {
}

func NewMysqlPreparation(config *core.MysqlConfig) *MysqlPreparation {
	p := MysqlPreparation{}
	return &p
}

func (p *MysqlPreparation) Prepare() {

}

const (
	MYSQL_PATH                  = "/deploy/database/mysql/"
	MYSQL_DB_Path               = MYSQL_PATH + "db_kube_killer.sql"
	MYSQL_TABLE_TERMINATOR_PATH = MYSQL_PATH + "table_terminator.sql"
	DEFAULT_DB_NAME             = "kube_killer"
)

func (p *MysqlPreparation) CreateDb(dbName string) error {
	content, err := ioutil.ReadFile(MYSQL_DB_Path)
	if err != nil {
		return err
	}
	if dbName == "" {
		dbName = DEFAULT_DB_NAME
	}
	db := p.GetDB()
	sql := strings.ReplaceAll(string(content), "@db", dbName)
	db.Exec(sql)
	return nil
}

func (m *MysqlPreparation) GetDB() *gorm.DB {
	db, err := gorm.Open("mysql", "user:password@tcp(192.168.1.4:3306)/dbname?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer db.Close()
	return db
}
