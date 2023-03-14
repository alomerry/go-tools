package utils

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"io/fs"
	"os"
	"time"
)

var (
	CWatcher *Watcher
)

type Watcher struct {
	conf   *config
	ticker *time.Ticker
}

type config struct {
	localPath  string
	remotePath string
	interval   time.Duration
}

func init() {
	CWatcher = &Watcher{
		conf: &config{
			localPath:  cast.ToString(viper.GetStringMap("local")["path"]),
			remotePath: cast.ToString(viper.GetStringMap("remote")["path"]),
			interval:   time.Second * time.Duration(viper.GetInt64("interval")),
		},
	}
	CWatcher.ticker = time.NewTicker(CWatcher.conf.interval)
}

func (w *Watcher) Watch() {
	w.ticker.Stop()
	w.ticker.Reset(CWatcher.conf.interval)
	for {
		<-w.ticker.C
		go w.Walk()
	}
}

func (w *Watcher) Walk() {
	fs.WalkDir(os.DirFS(w.conf.localPath), ".", upsertFile)
}

func upsertFile(relatePath string, d fs.DirEntry, err error) error {
	if d.IsDir() {
		return nil
	}
	//if needIgnore(relatePath) {
	//	return nil
	//}
	return nil
}
