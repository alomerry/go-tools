package dns

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	alidnsEndpoint string = "https://alidns.aliyuncs.com/"
)

// https://help.aliyun.com/document_detail/29776.html?spm=a2c4g.11186623.6.672.715a45caji9dMA
// Alidns Alidns
type Alidns struct {
	AK      string
	SK      string
	Domains []string
	NewAddr string
}

// AlidnsRecord record
type AlidnsRecord struct {
	DomainName string
	RecordID   string
	Value      string
}

// AlidnsSubDomainRecords 记录
type AlidnsSubDomainRecords struct {
	TotalCount    int
	DomainRecords struct {
		Record []AlidnsRecord
	}
}

// AlidnsResp 修改/添加返回结果
type AlidnsResp struct {
	RecordID  string
	RequestID string
}

func (ali *Alidns) UpsertDomainRecords() {
	for _, domain := range ali.Domains {
		var records AlidnsSubDomainRecords
		params := url.Values{}
		params.Set("Action", "DescribeSubDomainRecords")
		params.Set("SubDomain", domain)
		params.Set("Type", "A")
		err := ali.request(params, &records)

		if err != nil {
			panic(err)
		}

		if records.TotalCount > 0 {
			recordSelected := records.DomainRecords.Record[0]
			if params.Has("RecordId") {
				for i := 0; i < len(records.DomainRecords.Record); i++ {
					if records.DomainRecords.Record[i].RecordID == params.Get("RecordId") {
						recordSelected = records.DomainRecords.Record[i]
					}
				}
			}
			// 存在，更新
			ali.modify(recordSelected, domain)
		} else {
			// 不存在，创建
			ali.create(domain)
		}

	}
}

func (ali *Alidns) create(domain string) {
	params := url.Values{}
	params.Set("Action", "AddDomainRecord")
	params.Set("DomainName", strings.TrimPrefix(domain, strings.Split(domain, ".")[0]+"."))
	params.Set("RR", strings.Split(domain, ".")[0])
	params.Set("Type", "A")
	params.Set("Value", ali.NewAddr)
	params.Set("TTL", "600")

	var result AlidnsResp
	err := ali.request(params, &result)

	if err == nil && result.RecordID != "" {
		log.Printf("新增域名解析 %s 成功！IP: %s", domain, ali.NewAddr)
	} else {
		log.Printf("新增域名解析 %s 失败！", domain)
	}
}

func (ali *Alidns) modify(recordSelected AlidnsRecord, domain string) {
	if recordSelected.Value == ali.NewAddr {
		log.Printf("你的IP %s 没有变化, 域名 %s", ali.NewAddr, domain)
		return
	}

	params := url.Values{}
	params.Set("Action", "UpdateDomainRecord")
	params.Set("RR", strings.Split(domain, ".")[0])
	params.Set("RecordId", recordSelected.RecordID)
	params.Set("Type", "A")
	params.Set("Value", ali.NewAddr)
	params.Set("TTL", "600")

	var result AlidnsResp
	err := ali.request(params, &result)

	if err == nil && result.RecordID != "" {
		log.Printf("更新域名解析 %s 成功！IP: %s", domain, ali.NewAddr)
	} else {
		log.Printf("更新域名解析 %s 失败！", domain)
	}
}

// request 统一请求接口
func (ali *Alidns) request(params url.Values, result interface{}) (err error) {
	aliyunSigner(ali.AK, ali.SK, &params)

	req, err := http.NewRequest(
		"GET",
		alidnsEndpoint,
		bytes.NewBuffer(nil),
	)
	req.URL.RawQuery = params.Encode()

	if err != nil {
		log.Println("http.NewRequest失败. Error: ", err)
		return
	}

	resp, err := client.Do(req)
	err = getHTTPResponse(resp, alidnsEndpoint, err, result)

	return
}

func aliyunSigner(accessKeyID, accessSecret string, params *url.Values) {
	params.Set("SignatureMethod", "HMAC-SHA1")
	params.Set("SignatureNonce", strconv.FormatInt(time.Now().UnixNano(), 10))
	params.Set("AccessKeyId", accessKeyID)
	params.Set("SignatureVersion", "1.0")
	params.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	params.Set("Format", "JSON")
	params.Set("Version", "2015-01-09")
	params.Set("Signature", HmacSignToB64("HMAC-SHA1", "GET", accessSecret, *params))
}

func HmacSignToB64(signMethod string, httpMethod string, appKeySecret string, vals url.Values) (signature string) {
	return base64.StdEncoding.EncodeToString(HmacSign(signMethod, httpMethod, appKeySecret, vals))
}

var (
	signMethodMap = map[string]func() hash.Hash{
		"HMAC-SHA1":   sha1.New,
		"HMAC-SHA256": sha256.New,
		"HMAC-MD5":    md5.New,
	}
)

func HmacSign(signMethod string, httpMethod string, appKeySecret string, vals url.Values) (signature []byte) {
	key := []byte(appKeySecret + "&")

	var h hash.Hash
	if method, ok := signMethodMap[signMethod]; ok {
		h = hmac.New(method, key)
	} else {
		h = hmac.New(sha1.New, key)
	}
	makeDataToSign(h, httpMethod, vals)
	return h.Sum(nil)
}

type strToEnc struct {
	s string
	e bool
}

func makeDataToSign(w io.Writer, httpMethod string, vals url.Values) {
	in := make(chan *strToEnc)
	go func() {
		in <- &strToEnc{s: httpMethod}
		in <- &strToEnc{s: "&"}
		in <- &strToEnc{s: "/", e: true}
		in <- &strToEnc{s: "&"}
		in <- &strToEnc{s: vals.Encode(), e: true}
		close(in)
	}()

	specialUrlEncode(in, w)
}

var (
	encTilde = "%7E"         // '~' -> "%7E"
	encBlank = []byte("%20") // ' ' -> "%20"
	tilde    = []byte("~")
)

func specialUrlEncode(in <-chan *strToEnc, w io.Writer) {
	for s := range in {
		if !s.e {
			io.WriteString(w, s.s)
			continue
		}

		l := len(s.s)
		for i := 0; i < l; {
			ch := s.s[i]

			switch ch {
			case '%':
				if encTilde == s.s[i:i+3] {
					w.Write(tilde)
					i += 3
					continue
				}
				fallthrough
			case '*', '/', '&', '=':
				fmt.Fprintf(w, "%%%02X", ch)
			case '+':
				w.Write(encBlank)
			default:
				fmt.Fprintf(w, "%c", ch)
			}

			i += 1
		}
	}
}
