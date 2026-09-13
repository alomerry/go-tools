package analysis

import (
	"os"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestCustomAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	if _, err := os.Stat(testdata); err != nil {
		// testdata 目录缺失（.gitignore 曾全局忽略 testdata 导致从未入库）：
		// 需要含 testdata/src/redis 的样例代码与 // want 断言方可运行
		t.Skip("testdata not present, skipping analysistest")
	}
	// Run 函数会加载 testdata/src/redis 目录下的 Go 代码并执行 Analyzer
	analysistest.Run(t, testdata, CustomAnalyzer, "redis")
}
