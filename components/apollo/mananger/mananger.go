package mananger

import (
  "sync"
  
  apollo2 "github.com/alomerry/go-tools/model/apollo"
  "github.com/alomerry/go-tools/static/env"
  "github.com/alomerry/go-tools/static/env/influxdb"
  "github.com/alomerry/go-tools/utils/apollo"
)

var (
  manager     = &apolloCfgManager{}
	managerOnce sync.Once
)

func NewApolloManager() Manager {
  managerOnce.Do(func() {
		manager = &apolloCfgManager{}
	})

	return manager
}

type apolloCfgManager struct {
}

func (*apolloCfgManager) GetBarkCfg() *apollo2.BarkConfig {
  return apollo.GetBarkCfg()
}

func (*apolloCfgManager) GetMysqlConfig() *apollo2.MysqlConfig {
  return apollo.GetMysqlConfig()
}

func (*apolloCfgManager) GitCfg() *apollo2.GitConfig {
  return apollo.GitCfg()
}

func (*apolloCfgManager) GitToken(provider string) string {
  item, exists := apollo.GitCfg().Providers[provider]
  if exists {
    return item.Token
  }
  return ""
}

func (*apolloCfgManager) GetRedisConfig() *apollo2.RedisConfig {
  return apollo.GetRedisConfig()
}

func (*apolloCfgManager) InfluxDbConfig() *apollo2.InfluxDbConfig {
  return apollo.GetInfluxDbConfig()
}

func (*apolloCfgManager) InfluxOrg() string {
  if env.Local() {
    return influxdb.GetOrg()
  }
  return apollo.GetInfluxDbConfig().Org
}

func (*apolloCfgManager) KookCfg() *apollo2.KookConfig {
  return apollo.GetKookConfig()
}

func (*apolloCfgManager) InfluxToken() string {
  if env.Local() {
    return influxdb.GetToken()
  }
  return apollo.GetInfluxDbConfig().Token
}

func (*apolloCfgManager) InfluxEndpoint() string {
  if env.Local() {
    return influxdb.GetEndpoint()
  }
  return apollo.GetInfluxDbConfig().Endpoint
}

func (*apolloCfgManager) KafkaCfg() *apollo2.KafkaConfig {
  return apollo.GetKafkaConfig()
}
