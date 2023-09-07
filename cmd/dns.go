package cmd

import (
	gdns "github.com/alomerry/go-tools/dns"
	"github.com/spf13/cobra"
)

var (
	domainsFile string
)

var dns = &cobra.Command{
	Use:   "dns",
	Short: "dns tools help your to set domain A dns",
	Run: func(cmd *cobra.Command, args []string) {
		client := gdns.GenClient(domainsFile)
		client.SetDomainA()
	},
}

func init() {
	dns.Flags().StringVarP(&domainsFile, "config", "f", "./domains.toml", "domains list file")
	RootCmd.AddCommand(dns)
}
