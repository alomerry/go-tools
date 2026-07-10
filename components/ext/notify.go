package ext

import (
	"context"
	"sync"

	"github.com/alomerry/go-tools/components/notify"
	"github.com/alomerry/go-tools/static/cons"
	notify2 "github.com/alomerry/go-tools/static/cons/notify"
)

// NotifyTypeResolver maps a type name string to a NotifySenderType. Callers
// whose mapping comes from a business apollo key (e.g. backend's NotifyConfig)
// should inject it via SetNotifyTypeResolver before LoadExt.
type NotifyTypeResolver func(typeName string) notify2.NotifySenderType

var (
	defaultNotifyTypeResolver NotifyTypeResolver = func(typeName string) notify2.NotifySenderType {
		return notify2.ToNotifySenderType(typeName)
	}

	notifier   NotifyManager
	notifyOnce sync.Once
)

// SetNotifyTypeResolver overrides the typeName->senderType resolver used by
// NotifyExt.WithNotifierByName.
func SetNotifyTypeResolver(r NotifyTypeResolver) {
	if r != nil {
		defaultNotifyTypeResolver = r
	}
}

func init() {
	Register(cons.ExtNotify, NewNotifyExt)
}

type notifyExt struct {
	*notify.Manager
}

func NewNotifyExt() Ext {
	notifyOnce.Do(func() {
		notifier = &notifyExt{
			notify.NewManager(),
		}
	})
	return notifier
}

func Notifier() NotifyManager {
	return notifier
}

type NotifyManager interface {
	Ext

	Register(name notify2.NotifySenderType, notifier notify.Notifier)
	Unregister(name notify2.NotifySenderType)
	Get(name notify2.NotifySenderType) (notify.NotifierWrapper, bool)
	WithNotifier(ctx context.Context, senderType notify2.NotifySenderType) notify.NotifierWrapper
	WithNotifierByName(ctx context.Context, typeName string) notify.NotifierWrapper
	Send(ctx context.Context, name notify2.NotifySenderType, opts ...notify.Option) error
	Broadcast(ctx context.Context, opts ...notify.Option) map[notify2.NotifySenderType]error
	Close() error
}

func (n *notifyExt) Init(_ context.Context) error {
	return n.InitDrivers()
}

func (n *notifyExt) WithNotifierByName(ctx context.Context, typeName string) notify.NotifierWrapper {
	return n.WithNotifier(ctx, defaultNotifyTypeResolver(typeName))
}
