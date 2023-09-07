package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "go-tools",
	Short: "go tools help your do several things.",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		panic(err)
	}
}
