package db

import (
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/utils/db/mysql"
)

const (
	defaultDumpPath = "/tmp/db-dump"
)

type DumpTool struct {
	dbCfg     map[cons.DatabaseType]map[string]any
	clientMap map[cons.DatabaseType]innerDumpTool

	dumpPath string
}

type innerDumpTool interface {
	Dump(prefix string, param map[string]any, db cons.Database) (string, error)
}

type GenDumpCmdParamFunc func(*DumpTool)

func MySQLDumpCmdParam(user, host, port, password string) GenDumpCmdParamFunc {
	return func(tool *DumpTool) {
		tool.dbCfg[cons.MySQL] = map[string]any{
			"user":     user,
			"host":     host,
			"port":     port,
			"password": password,
		}
	}
}

func SetDumpPath(path string) GenDumpCmdParamFunc {
	return func(tool *DumpTool) {
		tool.dumpPath = path
	}
}

func NewDumpTool(paramFunc ...GenDumpCmdParamFunc) *DumpTool {
	tool := &DumpTool{
		dumpPath: defaultDumpPath,
		dbCfg:    map[cons.DatabaseType]map[string]any{},
	}

	tool.clientMap = map[cons.DatabaseType]innerDumpTool{
		cons.MySQL: &mysql.DumpTool{},
	}

	for i := range paramFunc {
		paramFunc[i](tool)
	}
	return tool
}

func (d *DumpTool) DumpDbs(dbs ...cons.Database) ([]string, error) {
	var result []string
	for i := range dbs {
		sql, err := d.dumpDatabase(dbs[i])
		if err != nil {
			return nil, err
		}
		result = append(result, sql)
	}
	return result, nil
}

func (d *DumpTool) dumpDatabase(db cons.Database) (string, error) {
	switch db.Type {
	case cons.MySQL:
		break
	default:
		return cons.EmptyStr, nil
	}

	return d.clientMap[db.Type].Dump(d.dumpPath, d.dbCfg[db.Type], db)
}
