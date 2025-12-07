package apollo

import (
	"reflect"
	"strings"
	"sync"

	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/utils"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/sirupsen/logrus"
)

type apolloListener struct {
	watchKeys map[string]valueItem
	lock      sync.RWMutex
}

type valueItem struct {
	valueType string
	value     reflect.Value
}

func newApolloListener() *apolloListener {
	return &apolloListener{
		watchKeys: make(map[string]valueItem),
	}
}

func (a *apolloListener) TryWatchKey(key, valueType string, valuePointer any) bool {
	if !strings.HasSuffix(key, ",dynamic") {
		return false
	}
	a.lock.Lock()
	defer a.lock.Unlock()

	a.watchKeys[strings.TrimSuffix(key, ",dynamic")] = valueItem{
		valueType: valueType,
		value:     reflect.ValueOf(valuePointer).Elem(),
	}
	return true
}

func (a *apolloListener) OnNewestChange(event *storage.FullChangeEvent) {
	// empty
}

func (a *apolloListener) OnChange(event *storage.ChangeEvent) {
	for key, change := range event.Changes {
		a.lock.RLock()
		val, ok := a.watchKeys[key]
		if !ok {
			continue
		}

		newValue, ok := change.NewValue.(string)
		if !ok {
			logrus.Errorf("new value is not string type, t: %v", reflect.TypeOf(change.NewValue))
			a.lock.RUnlock()
			continue
		}

		var err error

		// only support json
		switch val.valueType {
		case cons.ApolloValTypeJson:
			err = utils.SetJson(newValue, val.value)
		}

		if err != nil {
			logrus.Errorf("unmarshal new value failed, err: %v", err.Error())
			a.lock.RUnlock()
			continue
		}

		logrus.Infof("apollo config changed, key %s", key)
		a.lock.RUnlock()
	}
}
