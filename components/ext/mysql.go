package ext

import (
  "context"
  "fmt"
  "sync"
  "time"
  
  "github.com/alomerry/go-tools/components/log"
  "github.com/alomerry/go-tools/components/mysql"
  "github.com/alomerry/go-tools/static/cons"
  "github.com/alomerry/go-tools/static/env"
  "github.com/uptrace/bun"
  "github.com/uptrace/bun/extra/bundebug"
)

var (
  MySQLExt  *MySQLExtension
  mysqlOnce sync.Once
)

func init() {
  Register(cons.ExtMySQL, NewMySQLExtension)
}

type MySQLExtension struct {
  db *bun.DB
}

func NewMySQLExtension() Ext {
  mysqlOnce.Do(func() {
    MySQLExt = &MySQLExtension{}
  })
  return MySQLExt
}

// TODO https://bun.uptrace.dev/guide/performance-monitoring.html

func (m *MySQLExtension) Init(ctx context.Context) error {
  var err error
  m.db, err = mysql.NewBunMySqlClient(env.GetMysqlAdminDSN(Apollo().GetMysqlConfig().Uri))
  if err != nil {
    return fmt.Errorf("create mysql client failed: %w", err)
  }
  m.db.AddQueryHook(bundebug.NewQueryHook(
    bundebug.WithVerbose(Apollo().GetMysqlConfig().Verbose),
    bundebug.FromEnv("BUNDEBUG"),
  ))
  
  go func() {
    _ = m.Watch()
  }()
  
  return nil
}

func (m *MySQLExtension) Watch() error {
  ticker := time.NewTicker(time.Second * 30)
  defer ticker.Stop()
  for {
    select {
    case <-ticker.C:
      ctx, cancel := context.WithTimeout(context.TODO(), time.Second*30)
      err := m.db.PingContext(ctx)
      cancel()
      // 数据库瞬时抖动是常态：goroutine 内 panic 无法被调用方 recover，
      // 只降级日志，由调用方告警体系（日志采集）负责跟进。
      if err != nil {
        log.Errorf(ctx, "mysql ping failed %v", err.Error())
      }
    }
  }
}

func GetConn(ctx context.Context) bun.Conn {
  conn, err := MySQLExt.db.Conn(ctx)
  if err != nil {
    log.Panicf(ctx, "get mysql conn failed %v", err.Error())
  }
  return conn
}

func GetDB() *bun.DB {
  return MySQLExt.db
}
