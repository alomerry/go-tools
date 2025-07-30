package apollo

import (
	"fmt"
	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"
)

func initMetaCfg() {
	c := &config.AppConfig{
		AppID:          "homelab",
		Cluster:        "default",
		IP:             "https://apollo-cfg.alomerry.cn",
		IsBackupConfig: false,
		MustStart:      true,
		NamespaceName:  "application",
		Secret:         "7d8507b0eaad45579efa40d5153d727c",
	}

	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})

	if err != nil {
		panic(err)
	}

	//Use your apollo key to test
	cache := client.GetConfigCache(c.NamespaceName)
	value, err := cache.Get("backend.openapi.meta")
	fmt.Println(value, err)
	cache.Range(func(key, value interface{}) bool {
		fmt.Println("key : ", key, ", value :", value)
		return true
	})
}
