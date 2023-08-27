package main

import (
	"os"

	"github.com/alomerry/go-tools/dns"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	pflag.String("f", "./domains.toml", "please set the config abstract path")
	pflag.Parse()
	viper.BindPFlag("f", pflag.Lookup("f"))
}

func main() {
	viper.SetConfigFile(viper.GetString("f"))
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}
	var (
		CF_SK   = os.Getenv("CF_SK")
		CF_ZONE = os.Getenv("CF_ZONE")
		ALI_AK  = os.Getenv("ALI_AK")
		ALI_SK  = os.Getenv("ALI_SK")
	)
	if CF_SK != "" {
		cf := &dns.Cloudflare{
			Secret:  CF_SK,
			Domains: viper.GetStringSlice("domains"),
			ZoneId:  CF_ZONE,
			NewAddr: dns.GetIpv4AddrFromUrl(),
		}
		cf.UpsertDomainRecords()
	}

	if ALI_AK != "" && ALI_SK != "" {
		ali := &dns.Alidns{
			AK:      ALI_AK,
			SK:      ALI_SK,
			Domains: viper.GetStringSlice("domains"),
			NewAddr: dns.GetIpv4AddrFromUrl(),
		}
		ali.UpsertDomainRecords()
	}
}
