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
  
  barkDynamic     *apollo.Dynamic[apollo2.BarkConfig]
  influxDBDynamic *apollo.Dynamic[apollo2.InfluxDbConfig]
  gitDynamic      *apollo.Dynamic[apollo2.GitConfig]
  kafkaDynamic    *apollo.Dynamic[apollo2.KafkaConfig]
  kookDynamic     *apollo.Dynamic[apollo2.KookConfig]
  mysqlDynamic    *apollo.Dynamic[apollo2.MysqlConfig]
  redisDynamic    *apollo.Dynamic[apollo2.RedisConfig]
  
  ossDynamic *apollo.Dynamic[model.Config]
  ossOnce    sync.Once
}

func (a *apolloCfgManager) GetBarkCfg() *apollo2.BarkConfig {
  a.barkOnce.Do(func() {
    d, err := apollo.GetJson[apollo2.BarkConfig](apollo3.BarkConfig)
    if err != nil {
      logrus.Errorf("get bark config failed: %v", err)
      return
    }
    a.barkDynamic = d
  })
  if a.barkDynamic == nil {
    return nil
  }
  return a.barkDynamic.Load()
}

func (a *apolloCfgManager) GetMysqlConfig() *apollo2.MysqlConfig {
  a.mysqlInitOnce.Do(func() {
    d, err := apollo.GetJson[apollo2.MysqlConfig](apollo3.MysqlConfig)
    if err != nil {
      logrus.Panic(err)
    }
    a.mysqlDynamic = d
  })
  return a.mysqlDynamic.Load()
}

func (a *apolloCfgManager) GitCfg() *apollo2.GitConfig {
  a.gitCfgOnce.Do(func() {
    d, err := apollo.GetJson[apollo2.GitConfig](apollo3.GitConfig)
    if err != nil {
      logrus.Panic(err)
    }
    a.gitDynamic = d
  })
  return a.gitDynamic.Load()
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
    d, err := apollo.GetJson[apollo2.RedisConfig](apollo3.RedisConfig)
    if err != nil {
      logrus.Panic(err)
    }
    a.redisDynamic = d
  })
  return a.redisDynamic.Load()
}

func (a *apolloCfgManager) InfluxDbConfig() *apollo2.InfluxDbConfig {
  a.influxDBInitOnce.Do(func() {
    d, err := apollo.GetJson[apollo2.InfluxDbConfig](apollo3.InfluxdbConfig)
    if err != nil {
      logrus.Panicf("init influx-db config failed: %v", err)
    }
    a.influxDBDynamic = d
  })
  return a.influxDBDynamic.Load()
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
    d, err := apollo.GetJson[apollo2.KookConfig](apollo3.KookConfig)
    if err != nil {
      logrus.Errorf("get kook config failed: %v", err)
      return
    }
    a.kookDynamic = d
  })
  if a.kookDynamic == nil {
    return nil
  }
  return a.kookDynamic.Load()
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
    d, err := apollo.GetJson[apollo2.KafkaConfig](apollo3.KafkaConfig)
    if err != nil {
      logrus.Panic(err)
    }
    a.kafkaDynamic = d
  })
  return a.kafkaDynamic.Load()
}

func (a *apolloCfgManager) RustFs() *model.Config {
  a.ossOnce.Do(func() {
    d, err := apollo.GetJson[model.Config](apollo3.RustFsConfig)
    if err != nil {
      logrus.Panic(err)
    }
    a.ossDynamic = d
  })
  return a.ossDynamic.Load()
}

