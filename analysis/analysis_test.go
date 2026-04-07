package analysis

import (
  "testing"
  
  "golang.org/x/tools/go/analysis/analysistest"
)

func TestCustomAnalyzer(t *testing.T) {
  testdata := analysistest.TestData()
  // Run 函数会加载 testdata/src/a 目录下的 Go 代码并执行 Analyzer
  analysistest.Run(t, testdata, CustomAnalyzer, "redis")
	analysistest.Run(t, "/Users/alomerry/workspace/go/go-tools", CustomAnalyzer, "./...")
}
