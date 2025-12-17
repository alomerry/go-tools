package utils

import (
	"net"
)

func IsIPV4(ip string) bool {
	addr := net.ParseIP(ip)
	return addr != nil && addr.To4() != nil
}
