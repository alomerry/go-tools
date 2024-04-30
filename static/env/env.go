package env

import (
	"os"
	"strings"

	"github.com/alomerry/go-tools/static/constant"
)

func GetElasticSearchEndpoint() string {
	return os.Getenv(constant.ELASTICSEARCH_ENDPOINT)
}

func GetWithLocalElasticSearchEndpoint() string {
	if GetEnv() == constant.EnvLocal {
		return os.Getenv(constant.LOCAL_ELASTICSEARCH_ENDPOINT)
	}
	return GetElasticSearchEndpoint()
}

func GetElasticSearchAK() string {
	return os.Getenv(constant.ELASTICSEARCH_PASSWORD)
}

func Debug() bool {
	return os.Getenv(constant.DEBUG) == constant.DEBUG
}

func GetEnv() string {
	return os.Getenv(constant.ENV)
}

func GetKubeConfig() string {
	return os.Getenv(constant.KUBECONFIG)
}

func GetDBSalt() string {
	return os.Getenv(constant.DATABASE_SALT)
}

func GetJwtSecret() string {
	return os.Getenv(constant.JWT_SECRET)
}

func GetRedisDSN() string {
	return os.Getenv(constant.REDIS_DSN)
}

func GetWithLocalRedisDSN() string {
	if GetEnv() == constant.EnvLocal {
		return os.Getenv(constant.LOCAL_REDIS_DSN)
	}
	return GetRedisDSN()
}

func GetRedisAK() string {
	return os.Getenv(constant.REDIS_AK)
}

func GetWithLocalRedisAK() string {
	if GetEnv() == constant.EnvLocal {
		return os.Getenv(constant.LOCAL_REDIS_AK)
	}
	return GetRedisAK()
}

func GetMysqlAdminDSN() string {
	return os.Getenv(constant.MYSQL_ADMIN_DSN)
}

func GetWithLocalMysqlAdminDSN() string {
	if GetEnv() == constant.EnvLocal {
		return os.Getenv(constant.LOCAL_MYSQL_ADMIN_DSN)
	}
	return GetMysqlAdminDSN()
}

func GetRedisClusterDSN() []string {
	return strings.Split(os.Getenv(constant.REDIS_CLUSTER_DSN), ",")
}
