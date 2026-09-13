package test

import "os"

// IntegrationEnabled 报告集成测试是否显式开启。外部环境依赖（Apollo/MySQL/
// Mongo/Redis/Kafka/K8s/SSH/Kook/OSS 等）的测试以 GO_TOOLS_TEST_INTEGRATION=1
// 为总开关，默认跳过，保证 go test ./... 在无外部环境的机器与 CI 上绿色可跑。
func IntegrationEnabled() bool {
	return os.Getenv("GO_TOOLS_TEST_INTEGRATION") == "1"
}
