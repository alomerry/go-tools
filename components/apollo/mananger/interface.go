package mananger

import (
	apollo2 "github.com/alomerry/go-tools/model/apollo"
)

type Manager interface {
	GetBarkCfg() *apollo2.BarkConfig
	GetMysqlConfig() *apollo2.MysqlConfig
	GetRedisConfig() *apollo2.RedisConfig
	InfluxDbConfig() *apollo2.InfluxDbConfig
	InfluxOrg() string
	KookCfg() *apollo2.KookConfig
	InfluxToken() string
	InfluxEndpoint() string
	KafkaCfg() *apollo2.KafkaConfig
  GitCfg() *apollo2.GitConfig
  GitToken(provider string) string
}
