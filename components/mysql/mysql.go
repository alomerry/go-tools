package mysql

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

var (
	DefaultClient *bun.DB
	once          sync.Once
)

func InitDefaultClient(dsn string) {
	if DefaultClient != nil {
		return
	}

	once.Do(func() {
		var err error
		DefaultClient, err = NewBunMySqlClient(dsn)
		if err != nil {
			panic(err)
		}
	})
}

func NewBunMySqlClient(dsn string) (*bun.DB, error) {
	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	db := bun.NewDB(sqldb, mysqldialect.New())
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Hour)
	return db, nil
}
