package influxdb

import (
	"github.com/alomerry/go-tools/static/cons/tsdb/influxdb"
	"os"
)

func GetEndpoint() string {
	if v := os.Getenv(influxdb.Endpoint); len(v) > 0 {
		return v
	}
	return "http://localhost:8086"
}

func GetToken() string {
	if v := os.Getenv(influxdb.Token); len(v) > 0 {
		return v
	}
	return ""
}
