package apollo

import (
  "encoding/json"
  "fmt"
  "log"
  "reflect"
  "strings"
  "sync"
  
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

// GetJson loads a JSON config struct from the apollo cache and returns a
// Dynamic[T] snapshot handle. The caller holds the handle and calls Load() to
// read the current config — the returned *T is an immutable snapshot, so
// reading its fields is always race-free.
//
// When the key name carries the ",dynamic" suffix, GetJson also registers an
// OnChange callback that parses the new value into a fresh *T and Stores it
// atomically; subsequent Load() calls return the updated snapshot. A key
// without the suffix is loaded once and never hot-reloaded (the handle is
// still valid; Load always returns the initial snapshot).
//
// On a parse failure during OnChange the old snapshot is retained (fail-open
// to the last-known-good value). The initial load failure is returned as err;
// the caller decides whether to panic or degrade (Dynamic.Load will return nil
// until a successful fill).
func GetJson[T any](name string) (*Dynamic[T], error) {
	cache := client.GetConfigCache(env.ApolloNamespace())
	value, err := cache.Get(strings.TrimSuffix(toKey(client.clientId, name), ",dynamic"))
	if err != nil {
		return nil, err
	}

	d := &Dynamic[T]{}
	t := new(T)

	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), t); err != nil {
			return nil, err
		}
	default:
		log.Panicf("unsupported type %v", reflect.TypeOf(value))
	}

	d.Store(t)

	// Only ",dynamic" keys register an OnChange callback that hot-reloads the
	// config. Non-dynamic keys (mysql/redis/mongo/kafka...) keep the initial
	// snapshot for the process lifetime.
	// 注册 key 与初始读同口径（带 clientId 前缀）：OnChange 事件推送的是
	// 配置中心的实际 key（clientId.name），裸 name 永不匹配、回调永不触发。
	listener.TryWatchKey(toKey(client.clientId, name), func(newVal string) {
		nt := new(T)
		if err := json.Unmarshal([]byte(newVal), nt); err != nil {
			// Parse failure: keep the old snapshot (fail-open to last-known-good).
			return
		}
		d.Store(nt)
	})

	return d, nil
}

func toKey(clientId string, key string) string {
  return fmt.Sprintf("%s.%s", clientId, key)
}
