package mananger

import (
  "sync"
  
  "github.com/alomerry/go-tools/components/apollo"
  "github.com/alomerry/go-tools/model"
  apollo2 "github.com/alomerry/go-tools/model/apollo"
  apollo3 "github.com/alomerry/go-tools/static/cons/apollo"
  "github.com/alomerry/go-tools/static/env"
  "github.com/alomerry/go-tools/static/env/influxdb"
  "github.com/sirupsen/logrus"
)

var (
  manager = &apolloCfgManager{}
	managerOnce sync.Once
)

func ApolloManager() Manager {
  managerOnce.Do(func() {
		manager = &apolloCfgManager{}
	})

	return manager
}

type apolloCfgManager struct {
  barkOnce         sync.Once
  influxDBInitOnce sync.Once
  gitCfgOnce       sync.Once
  kafkaInitOnce    sync.Once
  kookInitOnce     sync.Once
  mysqlInitOnce    sync.Once
  redisInitOnce    sync.Once
  
  barkConfig     *apollo2.BarkConfig
  influxDBConfig *apollo2.InfluxDbConfig
  gitCfg         *apollo2.GitConfig
  kafkaConfig    *apollo2.KafkaConfig
  kookConfig     *apollo2.KookConfig
  mysqlConfig    *apollo2.MysqlConfig
  redisConfig    *apollo2.RedisConfig
  
  ossConfig *model.Config
  ossOnce   sync.Once
}

func (a *apolloCfgManager) GetBarkCfg() *apollo2.BarkConfig {
  a.barkOnce.Do(func() {
    a.barkConfig = &apollo2.BarkConfig{}
    err := apollo.GetJson(apollo3.BarkConfig, a.barkConfig)
    if err != nil {
      logrus.Errorf("get bark config failed: %v", err)
    }
  })
  return a.barkConfig
}

func (a *apolloCfgManager) GetMysqlConfig() *apollo2.MysqlConfig {
  a.mysqlInitOnce.Do(func() {
    a.mysqlConfig = &apollo2.MysqlConfig{}
    err := apollo.GetJson(apollo3.MysqlConfig, a.mysqlConfig)
    if err != nil {
      logrus.Panic(err)
    }
  })
  return a.mysqlConfig
}

func (a *apolloCfgManager) GitCfg() *apollo2.GitConfig {
  a.gitCfgOnce.Do(func() {
    a.gitCfg = &apollo2.GitConfig{}
    err := apollo.GetJson(apollo3.GitConfig, a.gitCfg)
    if err != nil {
      logrus.Panic(err)
    }
  })
  return a.gitCfg
}

func (a *apolloCfgManager) GitToken(provider string) string {
  item, exists := a.GitCfg().Providers[provider]
  if exists {
    return item.Token
  }
  return ""
}

func (a *apolloCfgManager) GetRedisConfig() *apollo2.RedisConfig {
  a.redisInitOnce.Do(func() {
    a.redisConfig = &apollo2.RedisConfig{}
    err := apollo.GetJson(apollo3.RedisConfig, a.redisConfig)
    if err != nil {
      logrus.Panic(err)
    }
  })
  return a.redisConfig
}

func (a *apolloCfgManager) InfluxDbConfig() *apollo2.InfluxDbConfig {
  a.influxDBInitOnce.Do(func() {
    a.influxDBConfig = &apollo2.InfluxDbConfig{}
    err := apollo.GetJson(apollo3.InfluxdbConfig, a.influxDBConfig)
    if err != nil {
      logrus.Panicf("init influx-db config failed: %v", err)
    }
  })
  return a.influxDBConfig
}

func (a *apolloCfgManager) InfluxOrg() string {
  if env.Local() {
    return influxdb.GetOrg()
  }
  return a.InfluxDbConfig().Org
}

func (a *apolloCfgManager) KookCfg() *apollo2.KookConfig {
  a.kookInitOnce.Do(func() {
    defer func() {
      if r := recover(); r != nil {
        logrus.Errorf("recover from panic: %v", r)
      }
    }()
    a.kookConfig = &apollo2.KookConfig{}
    err := apollo.GetJson(apollo3.KookConfig, a.kookConfig)
    if err != nil {
      logrus.Errorf("get kook config failed: %v", err)
    }
  })
  return a.kookConfig
}

func (a *apolloCfgManager) InfluxToken() string {
  if env.Local() {
    return influxdb.GetToken()
  }
  return a.InfluxDbConfig().Token
}

func (a *apolloCfgManager) InfluxEndpoint() string {
  if env.Local() {
    return influxdb.GetEndpoint()
  }
  return a.InfluxDbConfig().Endpoint
}

func (a *apolloCfgManager) KafkaCfg() *apollo2.KafkaConfig {
  a.kafkaInitOnce.Do(func() {
    a.kafkaConfig = &apollo2.KafkaConfig{}
    err := apollo.GetJson(apollo3.KafkaConfig, a.kafkaConfig)
    if err != nil {
      logrus.Panic(err)
    }
  })
  return a.kafkaConfig
}

func (a *apolloCfgManager) RustFs() *model.Config {
  a.ossOnce.Do(func() {
    a.ossConfig = &model.Config{}
    err := apollo.GetJson(apollo3.RustFsConfig, &a.ossConfig)
    if err != nil {
      logrus.Panic(err)
    }
  })
  
  return a.ossConfig
}

