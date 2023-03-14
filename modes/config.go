package modes

import (
	"fmt"
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

func initConfig(ctx context.Context) *config {
	config := initConfigFile()

	config.modes = viper.GetStringSlice("modes")
	for _, mode := range config.modes {
		var task Task
		switch mode {
		case share.MODE_PUSHER:
			task = pusher.Pusher{}
		case share.MODE_SYNCER:
			//task = syncer.Syncer{}
		}
		task.InitConfig()
		config.tasks = append(config.tasks, task)
	}
	return config
}

func initConfigFile() *config {
	var (
		rawPath    = viper.GetString("configPath")
		configPath = fmt.Sprintf("%s/%s", share.ExPath, strings.TrimPrefix(strings.TrimPrefix(rawPath, share.ExPath), "/"))
	)
	viper.SetConfigFile(configPath)
	err := viper.MergeInConfig()
	if err != nil {
		panic(err)
	}
	return &config{}
}
