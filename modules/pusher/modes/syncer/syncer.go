package syncer

import (
	"time"

	"github.com/alomerry/go-tools/modules/pusher/utils"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

type Syncer struct {
	watcher *utils.Watcher
}

type config struct {
	localPath  string
	remotePath string
	interval   time.Duration
}

func (s *Syncer) InitConfig() {
	conf := &config{
		localPath:  cast.ToString(viper.GetStringMap("syncer")["local-path"]),
		remotePath: cast.ToString(viper.GetStringMap("syncer")["remote-path"]),
		interval:   time.Second * time.Duration(cast.ToInt64(viper.GetStringMap("syncer")["interval"])),
	}
	s.watcher = utils.GenWatcher(conf)
}

func (s *Syncer) Run(ctx context.Context) error {
	// TODO
	return nil
}

func (s *Syncer) Done() bool {
	// TODO
	return true
}

func (c *config) GetLocalPath() string {
	return c.localPath
}

func (c *config) GetRemotePath() string {
	return c.remotePath
}
func (c *config) GetInterval() time.Duration {
	return c.interval
}
