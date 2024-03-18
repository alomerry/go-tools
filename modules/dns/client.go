package dns

import (
	"fmt"
	impl2 "github.com/alomerry/go-tools/modules/dns/impl"
	"github.com/alomerry/go-tools/modules/dns/utils"
	"github.com/spf13/viper"
	"os"
)

const (
	CF_SK   = "CF_SK"
	CF_ZONE = "CF_ZONE"
	ALI_AK  = "ALI_AK"
	ALI_SK  = "ALI_SK"

	IMPL_ALI = "IMPL_ALI"
	IMPL_CF  = "IMPL_CF"
)

type client struct {
	impls *[]string
}

type Proxy interface {
	initByConfig(string)
	SetDomainA()
}

func getImpls() []string {
	var result []string
	if os.Getenv("CF_SK") != "xxx" {
		result = append(result, IMPL_CF)
	}

	if os.Getenv("ALI_AK") != "xxx" && os.Getenv("ALI_SK") != "xxx" {
		result = append(result, IMPL_CF)
	}
	return result
}

func GenClient(config string) Proxy {
	c := &client{}
	c.initByConfig(config)
	return c
}

func (c *client) initByConfig(config string) {
	viper.SetConfigFile(config)
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}

	impls := getImpls()
	c.impls = &impls
}

func (c *client) SetDomainA() {
	for _, impl := range *c.impls {
		// TODO 抽出循环中
		switch impl {
		case IMPL_CF:
			cf := &impl2.Cloudflare{
				Secret:  os.Getenv(CF_SK),
				ZoneId:  os.Getenv(CF_ZONE),
				Domains: viper.GetStringSlice("domains"),
				NewAddr: utils.GetIpv4AddrFromUrl(),
			}
			cf.UpsertDomainRecords()
		case IMPL_ALI:
			ali := &impl2.Alidns{
				AK:      os.Getenv(ALI_AK),
				SK:      os.Getenv(ALI_SK),
				Domains: viper.GetStringSlice("domains"),
				NewAddr: utils.GetIpv4AddrFromUrl(),
			}
			ali.UpsertDomainRecords()
		default:
			panic(fmt.Sprintf("unsupport impl: [%s]\n", impl))
		}
	}
}
