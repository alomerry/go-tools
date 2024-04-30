package mysql

import (
	"github.com/alomerry/go-tools/static/env"
	m "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
	"time"
)

var (
	mysql *gorm.DB
	once  = sync.Once{}
)

func Instance(dsnGetter func() string) {
	once.Do(func() {
		var (
			dsn = dsnGetter()
			err error
		)

		mysql, err = gorm.Open(m.Open(dsn), &gorm.Config{
			// PrepareStmt: false, // https://gorm.io/zh_CN/docs/performance.html#缓存预编译语句
		})
		if err != nil {
			panic(err)
		}

		mysql = mysql.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8 auto_increment=1")
		sqlDB, err := mysql.DB()
		if err != nil {
			panic(err)
		}
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetConnMaxLifetime(time.Hour)
	})
}

func Session() *gorm.DB {
	if env.Debug() {
		return mysql.Debug().Session(&gorm.Session{})
	}
	return mysql.Session(&gorm.Session{})
}
