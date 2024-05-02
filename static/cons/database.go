package cons

import (
	"errors"
	"regexp"
)

const (
	DbSchemeMySQL = "mysql"

	DsnInvalid                   = "dsn invalid"
	DsnInvalidNoSuitableResolver = "no suitable resolver"
)

var (
	MySQL = newDatabaseType("mysql")
)

var (
	Umami      = newDatabase(MySQL, "umami")
	WalineBlog = newDatabase(MySQL, "waline_blog")

	// mysql://xxx:xxx@example.com:3306/xxx
	dsnReg = regexp.MustCompile(`([\w]+)://(.*):(.*)@(.*):([0-9]+)/([\w_]+)`)
)

type DatabaseType string

func newDatabaseType(name string) DatabaseType {
	return DatabaseType(name)
}

type Database struct {
	Type DatabaseType
	Name string
}

func newDatabase(dbType DatabaseType, name string) Database {
	return Database{dbType, name}
}

type BaseDbInfo struct {
	User     string
	Password string
	Host     string
	Port     string
}

func ParseDbDsn(dsn string) (*BaseDbInfo, error) {
	if !dsnReg.MatchString(dsn) {
		return nil, errors.New("不含 dsn")
	}

	res := dsnReg.FindStringSubmatch(dsn)
	if len(res) < 5 {
		return nil, errors.New(DsnInvalid)
	}
	switch res[1] {
	case DbSchemeMySQL:
		return &BaseDbInfo{
			User:     res[2],
			Password: res[3],
			Host:     res[4],
			Port:     res[5],
		}, nil
	default:
		return nil, errors.New(DsnInvalidNoSuitableResolver)
	}
	return nil, nil
}
