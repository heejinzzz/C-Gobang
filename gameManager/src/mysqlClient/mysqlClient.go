package mysqlClient

import (
	"database/sql"
	"gameManager/errorRecorder"
	_ "github.com/go-sql-driver/mysql"
)

const (
	mysqlMasterAddr = "C-Gobang-mysql-master:3306"
	mysqlSlaveAddr  = "C-Gobang-mysql-slave:3306"
)

var (
	MasterDB *sql.DB
	SlaveDB  *sql.DB
)

func init() {
	var err error
	// 连接 mysql-master
	MasterDB, err = sql.Open("mysql", "root:518315@tcp("+mysqlMasterAddr+")/C_Gobang")
	if err != nil {
		errorRecorder.RecordError("[userManager][open Mysql MasterDB failed][" + err.Error() + "]")
		panic(err)
	}
	err = MasterDB.Ping()
	if err != nil {
		errorRecorder.RecordError("[userManager][connect to Mysql MasterDB failed][" + err.Error() + "]")
		panic(err)
	}

	// 连接 mysql-slave
	SlaveDB, err = sql.Open("mysql", "root:518315@tcp("+mysqlSlaveAddr+")/C_Gobang")
	if err != nil {
		errorRecorder.RecordError("[userManager][open Mysql SlaveDB failed][" + err.Error() + "]")
		panic(err)
	}
	err = SlaveDB.Ping()
	if err != nil {
		errorRecorder.RecordError("[userManager][connect to Mysql SlaveDB failed][" + err.Error() + "]")
		panic(err)
	}
}
