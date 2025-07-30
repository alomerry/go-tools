package apollo

import (
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/sirupsen/logrus"
)

type apolloListener struct {
}

func (a *apolloListener) OnNewestChange(event *storage.FullChangeEvent) {
	//TODO implement me
	logrus.Warnf("TODO implement OnNewestChange")
}

func (a *apolloListener) OnChange(event *storage.ChangeEvent) {
	logrus.Printf("apollo config changed, namespace %s, %v", event.Namespace, event.Changes)
}
