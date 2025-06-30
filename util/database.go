package util

import (
	"fmt"
	"net/url"

	"github.com/devproje/neko-engine/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase() *Database {
	return &Database{}
}

func (d *Database) Open() error {
	cnf := config.Load().Database
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local",
		url.QueryEscape(cnf.Username),
		cnf.Password,
		cnf.URL, cnf.Port, cnf.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	d.db = db
	return nil
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}

func (d *Database) Close() error {
	if d.db != nil {
		sqlDB, err := d.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
