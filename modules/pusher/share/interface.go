package share

import "time"

type WatcherGetter interface {
	GetLocalPath() string
	GetRemotePath() string
	GetInterval() time.Duration
}
