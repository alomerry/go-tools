package main

import (
	"github.com/alomerry/go-tools/modules/pusher/modes"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

func init() {
	pflag.String("configPath", "./core.toml", "please set the config abstract path")
	pflag.Parse()
	viper.BindPFlag("configPath", pflag.Lookup("configPath"))
}

func main() {
	ctx := context.Background()

	// TODO add config validate before run
	err := modes.IClient.Init(ctx, "TODO").Run(ctx)

	if err != nil {
		panic(err)
	}
}
