package mysql

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

var (
	DefaultClient *bun.DB
	once          sync.Once
	initErr       error
)

// InitDefaultClient 幂等初始化包级默认连接，并发安全。原实现在 once 外裸读
// DefaultClient（race）且失败 panic 后 once 被消耗、后续调用永远拿到 nil。
// 现失败以 error 返回（DefaultClient 为 nil，不重试——连接配置错误应视为
// 致命问题由调用方终止启动）。
func InitDefaultClient(dsn string) error {
	once.Do(func() {
		DefaultClient, initErr = NewBunMySqlClient(dsn)
	})
	return initErr
}

func NewBunMySqlClient(dsn string) (*bun.DB, error) {
	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql failed: %w", err)
	}

	db := bun.NewDB(sqldb, mysqldialect.New())
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Hour)
	return db, nil
}
