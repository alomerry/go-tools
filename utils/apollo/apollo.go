package apollo

import (
  "sync"
  
  apollo2 "github.com/alomerry/go-tools/components/apollo"
  "github.com/alomerry/go-tools/model/apollo"
  apollo3 "github.com/alomerry/go-tools/static/cons/apollo"
  "github.com/sirupsen/logrus"
)

var (
  barkConfig = &apollo.BarkConfig{}
  influxDBConfig   = &apollo.InfluxDbConfig{}
  gitCfg     = &apollo.GitConfig{}
  kafkaConfig   = &apollo.KafkaConfig{}
  kookConfig   = &apollo.KookConfig{}
  mysqlConfig = &apollo.MysqlConfig{}
  redisConfig   = &apollo.RedisConfig{}
  
  barkOnce   sync.Once
  influxDBInitOnce sync.Once
  gitCfgOnce sync.Once
  kafkaInitOnce sync.Once
  kookInitOnce sync.Once
  mysqlInitOnce sync.Once
  redisInitOnce sync.Once
)

func GetBarkCfg() *apollo.BarkConfig {
  barkOnce.Do(func() {
    err := apollo2.GetJson(apollo3.BarkConfig, barkConfig)
    if err != nil {
      logrus.Errorf("get bark config failed: %v", err)
    }
  })
  return barkConfig
}

func GitCfg() *apollo.GitConfig {
  gitCfgOnce.Do(func() {
    err := apollo2.GetJson(apollo3.GitConfig, gitCfg)
    if err != nil {
      logrus.Panic(err)
    }
  })
  return gitCfg
}

func GetInfluxDbConfig() *apollo.InfluxDbConfig {
  influxDBInitOnce.Do(func() {
    err := apollo2.GetJson(apollo3.InfluxdbConfig, influxDBConfig)
    if err != nil {
      logrus.Panicf("init influx-db config failed: %v", err)
    }
  })
  return influxDBConfig
}

func GetKafkaConfig() *apollo.KafkaConfig {
  kafkaInitOnce.Do(func() {
    err := apollo2.GetJson(apollo3.KafkaConfig, kafkaConfig)
    if err != nil {
      logrus.Panic(err)
    }
  })
  return kafkaConfig
}

func GetKookConfig() *apollo.KookConfig {
  kookInitOnce.Do(func() {
    defer func() {
      if r := recover(); r != nil {
        logrus.Errorf("recover from panic: %v", r)
      }
    }()
    err := apollo2.GetJson(apollo3.KookConfig, kookConfig)
    if err != nil {
      logrus.Errorf("get kook config failed: %v", err)
    }
  })
  return kookConfig
}

func GetMysqlConfig() *apollo.MysqlConfig {
  mysqlInitOnce.Do(func() {
    err := apollo2.GetJson(apollo3.MysqlConfig, mysqlConfig)
    if err != nil {
      logrus.Panic(err)
    }
  })
  return mysqlConfig
}

func GetRedisConfig() *apollo.RedisConfig {
  redisInitOnce.Do(func() {
    err := apollo2.GetJson(apollo3.RedisConfig, redisConfig)
    if err != nil {
      logrus.Panic(err)
    }
  })
  return redisConfig
}

