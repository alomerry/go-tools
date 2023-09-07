package utils

import (
	"github.com/alomerry/go-tools/dns/internal"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"
)

func GetIpv4AddrFromUrl() string {
	client := &http.Client{Timeout: 30 * time.Second, Transport: internal.NoProxyTcp4Transport}
	resp, err := client.Get("http://whatismyip.akamai.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	result := regexp.MustCompile(`((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])`).FindString(string(body))
	if result == "" {
		log.Printf("获取 IPv4 结果失败！返回值：%s\n", result)
	}
	return result
}
