package env

import "os"

func GetElasticSearchEndpoint() string {
	return os.Getenv(ELASTICSEARCH_ENDPOINT)
}

func GetElasticSearchAK() string {
	return os.Getenv(ELASTICSEARCH_PASSWORD)
}

func GetEnv() string {
	return os.Getenv(ENV)
}

func GetKubeConfig() string {
	return os.Getenv(KUBECONFIG)
}

func GetDBSalt() string {
	return os.Getenv(DATABASE_SALT)
}

func GetJwtSecret() string {
	return os.Getenv(JWT_SECRET)
}

func GetRedisDSN() string {
	return os.Getenv(REDIS_DSN)
}

func GetRedisAK() string {
	return os.Getenv(REDIS_AK)
}

func GetMysqlAdminDSN() string {
	return os.Getenv(MYSQL_ADMIN_DSN)
}
