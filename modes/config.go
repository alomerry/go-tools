package modes

import (
	"fmt"
	"github.com/alomerry/go-pusher/component/oss"
	"github.com/alomerry/go-pusher/modes/pusher"
	"github.com/alomerry/go-pusher/share"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"strings"
)

type config struct {
	modes []string

	tasks []Task
}

func initConfig(_ context.Context) *config {
	conf := initConfigFile()

	conf.modes = viper.GetStringSlice("modes")
	for _, mode := range conf.modes {
		var task Task
		switch mode {
		case share.MODE_PUSHER:
			oss.InitOSS()
			task = &pusher.Pusher{}
		case share.MODE_SYNCER:
			//task = &syncer.Syncer{}
		}
		task.InitConfig()
		conf.tasks = append(conf.tasks, task)
	}
	return conf
}

func initConfigFile() *config {
	var (
		rawPath    = viper.GetString("configPath")
		configPath string
	)

	configPath = fmt.Sprintf("%s/%s", share.ExPath, strings.TrimPrefix(strings.TrimPrefix(rawPath, share.ExPath), "/"))

	viper.SetConfigFile(configPath)
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}
	return &config{}
}
