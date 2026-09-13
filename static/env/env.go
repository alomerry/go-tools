package env

import (
	"os"
	"strings"

	"github.com/alomerry/go-tools/static/cons"
)

func GetService() string {
	return os.Getenv(cons.Service)
}

func GetElasticSearchEndpoint(defaultVal ...string) string {
	if Local() {
		return os.Getenv(cons.LOCAL_ELASTICSEARCH_ENDPOINT)
	}

	if len(defaultVal) > 0 && len(defaultVal[0]) > 0 {
		return defaultVal[0]
	}

	return os.Getenv(cons.ELASTICSEARCH_ENDPOINT)
}

func GetCloudflareAccountId() string {
	return os.Getenv(cons.CloudflareAccountId)
}

func GetCloudflareR2AccountSK() string {
	return os.Getenv(cons.CloudflareR2AccountSK)
}

func GetCloudflareR2AccountKey() string {
	return os.Getenv(cons.CloudflareR2AccountKey)
}

func GetRustfsEndpoint() string {
	return os.Getenv(cons.RustFsEndpoint)
}

func GetRustfsAccessKey() string {
	return os.Getenv(cons.RustFsAccessKey)
}

func GetRustfsSecretKey() string {
	return os.Getenv(cons.RustFsSecretKey)
}

// GetElasticSearchAK 优先读 ELASTICSEARCH_API_KEY，兼容旧变量名
// ELASTICSEARCH_PASSWORD（历史命名与语义不符）
func GetElasticSearchAK() string {
	if v := os.Getenv("ELASTICSEARCH_API_KEY"); v != "" {
		return v
	}
	return os.Getenv(cons.ELASTICSEARCH_PASSWORD)
}

// Debug 宽松解析 DEBUG 环境变量：1/true/debug（大小写不敏感）均视为开启
// （原实现要求 DEBUG=DEBUG 精确匹配，true/1 不生效）
func Debug() bool {
	switch strings.ToLower(os.Getenv(cons.DEBUG)) {
	case "1", "true", "debug":
		return true
	}
	return false
}

func Local() bool {
	return GetEnv() == cons.EnvLocal
}

func GetEnv() string {
	return os.Getenv(cons.ENV)
}

func GetKubeConfig() string {
	return os.Getenv(cons.KUBECONFIG)
}

func GetDBSalt() string {
	return os.Getenv(cons.DATABASE_SALT)
}

func GetJwtSecret() string {
	return os.Getenv(cons.JWT_SECRET)
}

func GetRedisDSN(defaultVal ...string) string {
	if Local() {
		return os.Getenv(cons.LOCAL_REDIS_DSN)
	}

	if len(defaultVal) > 0 && len(defaultVal[0]) > 0 {
		return defaultVal[0]
	}

	return os.Getenv(cons.REDIS_DSN)
}

func GetRedisAK(defaultVal ...string) string {
	if Local() {
		return os.Getenv(cons.LOCAL_REDIS_AK)
	}

	if len(defaultVal) > 0 && len(defaultVal[0]) > 0 {
		return defaultVal[0]
	}

	return os.Getenv(cons.REDIS_AK)
}

func GetMysqlAdminDSN(defaultVal ...string) string {
	if Local() {
    return os.Getenv(cons.MYSQL_ADMIN_DSN)
	}

	if len(defaultVal) > 0 && len(defaultVal[0]) > 0 {
		return defaultVal[0]
	}
	return os.Getenv(cons.MYSQL_ADMIN_DSN)
}

func GetRedisClusterDSN() []string {
	return strings.Split(os.Getenv(cons.REDIS_CLUSTER_DSN), ",")
}

func GetKafkaUserName() string {
	return os.Getenv(cons.KafkaUser)
}

func GetKafkaPassword() string {
	return os.Getenv(cons.KafkaPassword)
}
