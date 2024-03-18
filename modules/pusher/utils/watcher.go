package utils

import (
	"io/fs"
	"os"
	"time"

	"github.com/alomerry/go-tools/pusher/share"
)

type Watcher struct {
	share.WatcherGetter
	ticker *time.Ticker
}

func GenWatcher(conf share.WatcherGetter) *Watcher {
	return &Watcher{conf, time.NewTicker(conf.GetInterval())}
}

func (w *Watcher) Watch() {
	w.ticker.Stop()
	w.ticker.Reset(w.GetInterval())
	for {
		<-w.ticker.C
		go w.Walk()
	}
}

func (w *Watcher) Walk() {
	fs.WalkDir(os.DirFS(w.GetLocalPath()), ".", upsertFile)
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
