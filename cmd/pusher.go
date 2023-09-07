package cmd

import (
	"context"
	ppusher "github.com/alomerry/go-tools/pusher/modes"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

var pusher = &cobra.Command{
	Use:   "pusher",
	Short: "TODO",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// TODO add config validate before run
		err := ppusher.IClient.Init(ctx, configPath).Run(ctx)

		if err != nil {
			panic(err)
		}
	},
}

func init() {
	pusher.Flags().StringVarP(&configPath, "config", "f", "./core.toml", "set the config abstract path")
	RootCmd.AddCommand(pusher)
}
