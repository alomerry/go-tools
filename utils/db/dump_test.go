package db

import (
	"fmt"
	"github.com/alomerry/go-tools/static/constant"
	"testing"
)

func TestDumpDatabase(t *testing.T) {
	tool := NewDumpTool(
		MySQLDumpCmdParam("root", "alomerry.com", "13306", "e0t=ereFqvpibm}91Y:n"),
		SetDumpPath("/Users/alomerry/workspace/go-tools/output"),
	)
	fmt.Println(tool.DumpDbs(constant.WalineBlog))
}
