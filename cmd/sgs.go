package cmd

import (
	"github.com/alomerry/go-tools/sgs/delay"
	"github.com/alomerry/go-tools/sgs/tools"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

const (
	MERGE_EXCEL = iota + 1
	DELAY_SUMMARY
	DELAY_REASON
)

var module string

var sgs = &cobra.Command{
	Use:   "sgs",
	Short: "sgs tools help your do something",
	Run: func(cmd *cobra.Command, args []string) {
		switch cast.ToInt(module) {
		case MERGE_EXCEL:
			tools.DoMergeExcelSheets()
		case DELAY_SUMMARY:
			delay.DoDelaySummaryMulti("")
		case DELAY_REASON:
			delay.DoDelayReason()
		}
	},
}

func init() {
	sgs.Flags().StringVarP(&module, "module", "m", "", "sgs 模块包含三个功能，请使用 sgs -m <数字> 来选择执行的任务：\n1. 合并表格\n2. delay 月报\n3. 广州 delay")
	sgs.MarkFlagRequired("module")
	RootCmd.AddCommand(sgs)
}
