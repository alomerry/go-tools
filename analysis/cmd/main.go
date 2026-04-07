package main

import (
  "github.com/alomerry/go-tools/analysis"
  "golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
  singlechecker.Main(analysis.CustomAnalyzer)
}
