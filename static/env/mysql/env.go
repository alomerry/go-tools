package mysql

import (
	"os"

	"github.com/alomerry/go-tools/static/cons/mysql"
)

func GetEndpoint() string {
	if v := os.Getenv(mysql.Endpoint); len(v) > 0 {
		return v
	}
	return "localhost:3306"
}

func GetUsername() string {
	if v := os.Getenv(mysql.Username); len(v) > 0 {
		return v
	}
	return ""
}

func GetPassword() string {
	if v := os.Getenv(mysql.Password); len(v) > 0 {
		return v
	}
	return ""
}
