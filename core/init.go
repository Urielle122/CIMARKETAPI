package core

import (
	"database/sql"
	"sync"

	"services/config/database"
	logs "services/log"
)

var (
	once      sync.Once
	MysqlDb   *sql.DB
	ErroMysql error
)

func InitConnection() {
	logs.Init()
	once.Do(func() {
		MysqlDb, ErroMysql = database.ConnectToMySQL()
	})
}

func GetDB() *sql.DB {
	return MysqlDb
}
