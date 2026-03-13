package apollo

import (
  "encoding/json"
  "fmt"
  "log"
  "reflect"
  "strings"
  "sync"
  
  "github.com/alomerry/go-tools/static/cons"
  "github.com/alomerry/go-tools/static/env"
  "github.com/apolloconfig/agollo/v4"
  "github.com/apolloconfig/agollo/v4/env/config"
)

var (
  initOnce = sync.Once{}
  
  client   apolloClient
  listener *apolloListener
)

type apolloClient struct {
  agollo.Client
  clientId string
}

func Init(appId, clientId string) {
  initOnce.Do(func() {
    c := &config.AppConfig{
      AppID:          appId,
      Cluster:        env.ApolloCluster(),
      IP:             env.ApolloHost(),
      IsBackupConfig: false,
      MustStart:      true,
      NamespaceName:  env.ApolloNamespace(),
      Secret:         env.ApolloSK(),
    }
    
    cli, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
      return c, nil
    })
    
    if err != nil {
      log.Panicf("init apollo failed, err: %v", err)
    }
    client = apolloClient{cli, clientId}
    
    listener = newApolloListener()
    client.AddChangeListener(listener)
  })
}

func Get(name string) (any, error) {
  cache := client.GetConfigCache(env.ApolloNamespace())
  value, err := cache.Get(toKey(client.clientId, name))
  return value, err
}

func GetJson[T any](name string, dist *T) error {
	cache := client.GetConfigCache(env.ApolloNamespace())
  value, err := cache.Get(strings.TrimSuffix(toKey(client.clientId, name), ",dynamic"))
	if err != nil {
		return err
	}

	switch value.(type) {
	case string:
		err = json.Unmarshal([]byte(value.(string)), dist)
	default:
    log.Panicf("unsupported type %v", reflect.TypeOf(value))
  }
  
  if err != nil {
    return err
  }
  
  _ = listener.TryWatchKey(name, cons.ApolloValTypeJson, dist)
  
  return nil
}

func toKey(clientId string, key string) string {
  return fmt.Sprintf("%s.%s", clientId, key)
}
