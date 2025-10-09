package ua

import (
	"fmt"

	"github.com/mileusna/useragent"
)

type UA struct {
	Name   string
	OS     string
	Device string
}

func ParseUA(s string) UA {
	ua := useragent.Parse(s)
	return UA{
		Name:   fmt.Sprintf("%s v%s", ua.Name, ua.Version),
		OS:     fmt.Sprintf("%s v%s", ua.OS, ua.OSVersion),
		Device: ua.Device,
	}
}
