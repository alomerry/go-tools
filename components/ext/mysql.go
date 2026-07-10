package ext

import (
  "context"
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
    log.Panicf(ctx, "create mysql client failed %v", err.Error())
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
      if err != nil {
        log.Panicf(ctx, "mysql ping failed %v", err.Error())
      }
      cancel()
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
