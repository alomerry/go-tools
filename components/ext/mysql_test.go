package ext

import (
	"context"
	"testing"
	"time"

	"github.com/alomerry/go-tools/components/mysql"
	"github.com/alomerry/go-tools/static/env"
	"github.com/alomerry/go-tools/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MySQLExtSuite 测试包级函数 GetDB/GetConn。遵循项目的集成测试约定。全局单例
// MySQLExt 由手动构造，注入真实的 *bun.DB 并绕过 Init() —— Init() 会启动一个永不退出
// 的 Watch() goroutine，会在测试进程中泄漏；同时它会从 apollo 拉取配置，这两点对测试
// GetDB/GetConn 都无必要。当本机没有 MySQL 时，使用 `go test -short` 可跳过。
func TestMySQLExtSuite(t *testing.T) {
	suite.Run(t, new(MySQLExtSuite))
}

type MySQLExtSuite struct {
	test.BaseSuite
	// 保留一个句柄以便在 TearDownSuite 中关闭 db；MySQLExt 本身没有暴露 Close，
	// 且全局变量会在 SetupSuite 中被覆盖。
	db closedb
}

type closedb interface {
	Close() error
}

func (s *MySQLExtSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping mysql integration test in short mode")
	}

	db, err := mysql.NewBunMySqlClient(env.GetMysqlAdminDSN())
	if err != nil {
		s.T().Fatalf("connect mysql failed: %v", err)
	}

	// 手动将真实客户端注入全局单例。GetDB/GetConn 从该全局变量读取，因此必须先赋值。
	// 有意不调用 Init()，以避免 Watch() goroutine 泄漏。
	MySQLExt = &MySQLExtension{db: db}
	s.db = db
}

func (s *MySQLExtSuite) TearDownSuite() {
	if s.db != nil {
		// Close 尽力而为；bun.DB.Close 会委托给底层的 sql.DB。
		_ = s.db.Close()
	}
}

// TestGetDB 验证 GetDB 返回注入的 *bun.DB 且连接可达。
func (s *MySQLExtSuite) TestGetDB() {
	db := GetDB()
	assert.NotNil(s.T(), db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	assert.NoError(s.T(), db.PingContext(ctx))
}

// TestGetConn 验证 GetConn 返回可用的 bun.Conn。
func (s *MySQLExtSuite) TestGetConn() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn := GetConn(ctx)
	assert.NotNil(s.T(), conn)

	// 新获取的 conn 应当能执行一条简单查询。
	rows, err := conn.QueryContext(ctx, "SELECT 1")
	assert.NoError(s.T(), err)
	if rows != nil {
		_ = rows.Close()
	}
}
