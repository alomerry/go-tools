package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	zonesAPI string = "https://api.cloudflare.com/client/v4/zones"
)

// Cloudflare Cloudflare实现
type Cloudflare struct {
	Secret  string
	Domains []string
	ZoneId  string
	NewAddr string
}

// CloudflareZonesResp cloudflare zones返回结果
type CloudflareZonesResp struct {
	CloudflareStatus
	Result []struct {
		ID     string
		Name   string
		Status string
		Paused bool
	}
}

// CloudflareRecordsResp records
type CloudflareRecordsResp struct {
	CloudflareStatus
	Result []CloudflareRecord `json:"result"`
}

// CloudflareRecord 记录实体
type CloudflareRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

// CloudflareStatus 公共状态
type CloudflareStatus struct {
	Success  bool
	Messages []string
}

func (cf *Cloudflare) UpsertDomainRecords() {
	for _, domain := range cf.Domains {
		var records CloudflareRecordsResp
		err := cf.request(
			"GET",
			fmt.Sprintf(zonesAPI+"/%s/dns_records?name=%s&per_page=50", cf.ZoneId, domain),
			nil,
			&records,
		)

		if err != nil || !records.Success || len(records.Result) > 1 {
			log.Printf("域名解析 %s 失败！", domain)
			return
		}

		if len(records.Result) > 0 {
			// 更新
			cf.modify(records.Result[0], cf.ZoneId, domain, cf.NewAddr)
		} else {
			// 新增
			cf.create(cf.ZoneId, domain)
		}
	}
}

func (cf *Cloudflare) create(zoneID string, domain string) {
	record := &CloudflareRecord{
		Type:    "A",
		Name:    domain,
		Content: cf.NewAddr,
		Proxied: false,
		TTL:     60,
	}
	var status CloudflareStatus
	err := cf.request(
		"POST",
		fmt.Sprintf(zonesAPI+"/%s/dns_records", zoneID),
		record,
		&status,
	)
	if err == nil && status.Success {
		log.Printf("新增域名解析 %s 成功！IP: %s", domain, cf.NewAddr)
	} else {
		log.Printf("新增域名解析 %s 失败！Messages: %s", domain, status.Messages)
	}
}

func (cf *Cloudflare) modify(record CloudflareRecord, zoneID string, domain string, newAddr string) {
	if record.Content == newAddr {
		log.Printf("你的 IP %s 没有变化, 域名 %s", newAddr, domain)
		return
	}
	var status CloudflareStatus
	record.Content = newAddr
	record.TTL = 60
	record.Proxied = false
	err := cf.request(
		"PUT",
		fmt.Sprintf(zonesAPI+"/%s/dns_records/%s", zoneID, record.ID),
		record,
		&status,
	)
	if err == nil && status.Success {
		log.Printf("更新域名解析 %s 成功！IP: %s", domain, newAddr)
	} else {
		log.Printf("更新域名解析 %s 失败！Messages: %s", domain, status.Messages)
	}
}

// request 统一请求接口
func (cf *Cloudflare) request(method string, url string, data interface{}, result interface{}) (err error) {
	jsonStr := make([]byte, 0)
	if data != nil {
		jsonStr, _ = json.Marshal(data)
	}
	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		log.Println("http.NewRequest失败. Error: ", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+cf.Secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	err = getHTTPResponse(resp, url, err, result)

	return
}
