package env

import (
	"os"
	"strings"

	"github.com/alomerry/go-tools/static/cons"
)

func GetElasticSearchEndpoint() string {
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

func GetWithLocalElasticSearchEndpoint() string {
	if GetEnv() == cons.EnvLocal {
		return os.Getenv(cons.LOCAL_ELASTICSEARCH_ENDPOINT)
	}
	return GetElasticSearchEndpoint()
}

func GetElasticSearchAK() string {
	return os.Getenv(cons.ELASTICSEARCH_PASSWORD)
}

func Debug() bool {
	return os.Getenv(cons.DEBUG) == cons.DEBUG
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

func GetRedisDSN() string {
	return os.Getenv(cons.REDIS_DSN)
}

func GetWithLocalRedisDSN() string {
	if GetEnv() == cons.EnvLocal {
		return os.Getenv(cons.LOCAL_REDIS_DSN)
	}
	return GetRedisDSN()
}

func GetRedisAK() string {
	return os.Getenv(cons.REDIS_AK)
}

func GetWithLocalRedisAK() string {
	if GetEnv() == cons.EnvLocal {
		return os.Getenv(cons.LOCAL_REDIS_AK)
	}
	return GetRedisAK()
}

func GetMysqlAdminDSN() string {
	return os.Getenv(cons.MYSQL_ADMIN_DSN)
}

func GetWithLocalMysqlAdminDSN() string {
	if GetEnv() == cons.EnvLocal {
		return os.Getenv(cons.LOCAL_MYSQL_ADMIN_DSN)
	}
	return GetMysqlAdminDSN()
}

func GetRedisClusterDSN() []string {
	return strings.Split(os.Getenv(cons.REDIS_CLUSTER_DSN), ",")
}
