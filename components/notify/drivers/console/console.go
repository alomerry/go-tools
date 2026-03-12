package console

import (
	"context"

	"github.com/alomerry/go-tools/components/log"
  "github.com/alomerry/go-tools/components/notify"
  notify2 "github.com/alomerry/go-tools/static/cons/notify"
)

func init() {
	notify.Register(notify2.NotifySenderConsole, &Driver{})
}

type Driver struct{}

func (d *Driver) Open() (notify.Notifier, error) {
	return &Notifier{}, nil
}

type Notifier struct{}

func (n *Notifier) Send(ctx context.Context, msg *notify.Message) error {
	log.Infof(ctx, "[Console Notify] Subject: %s, Content: %s", msg.Subject, msg.Content)
	return nil
}

func (n *Notifier) Close() error {
	return nil
}
