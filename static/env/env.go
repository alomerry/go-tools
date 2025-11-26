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

func GetElasticSearchAK() string {
	return os.Getenv(cons.ELASTICSEARCH_PASSWORD)
}

func Debug() bool {
	return os.Getenv(cons.DEBUG) == cons.DEBUG
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
		return os.Getenv(cons.LOCAL_MYSQL_ADMIN_DSN)
	}

	if len(defaultVal) > 0 && len(defaultVal[0]) > 0 {
		return defaultVal[0]
	}
	return os.Getenv(cons.MYSQL_ADMIN_DSN)
}

func GetRedisClusterDSN() []string {
	return strings.Split(os.Getenv(cons.REDIS_CLUSTER_DSN), ",")
}
