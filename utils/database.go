package utils

import (
	"database/sql"
	"fmt"

	"github.com/devproje/neko-engine/config"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

type Database struct{}

func NewDatabase() *Database {
	return &Database{}
}

func (*Database) Open() (*sql.DB, error) {
	cnf := config.Load()
	datasource := mysql.NewConfig()
	datasource.Net = "tcp"
	datasource.DBName = cnf.Database.Name
	datasource.User = cnf.Database.Username
	datasource.Passwd = cnf.Database.Password
	datasource.Addr = fmt.Sprintf("%s:%d", cnf.Database.URL, cnf.Database.Port)
	datasource.ParseTime = true

	db, err := sql.Open("mysql", datasource.FormatDSN())
	if err != nil {
		return nil, err
	}

	return db, nil
}
