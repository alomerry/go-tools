package influxdb

import (
	"os"

	"github.com/alomerry/go-tools/static/cons/tsdb/influxdb"
)

func GetEndpoint() string {
	if v := os.Getenv(influxdb.Endpoint); len(v) > 0 {
		return v
	}
	return "http://localhost:8086"
}

func GetOrg() string {
	if v := os.Getenv(influxdb.Org); len(v) > 0 {
		return v
	}
	return "alomerry.com"
}

func GetToken() string {
	if v := os.Getenv(influxdb.Token); len(v) > 0 {
		return v
	}
	return ""
}
