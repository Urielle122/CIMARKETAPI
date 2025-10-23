package core 


import (
	"database/sql"
	"sync"

	"services/config/database"
	"services/logs"

)

var (
	once sync.Once
	MysqlDb *sql.DB
	ErroMysql error
)

func InitConnection(){
	logs.Init()
	once.Do(func() {
        MysqlDb, ErroMysql = database.ConnectToMySQL()
    })
}