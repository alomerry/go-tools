#!/bin/bash

# 将 components/collect/agent_demo/main.go 构建成 linux 二进制到 bin 目录
mkdir -p bin
rm -rf bin/agent_demo
GOOS=linux GOARCH=amd64 go build -o bin/agent_demo components/collect/agent_demo/main.go

#

cp bin/agent_demo /tmp/agent_demo