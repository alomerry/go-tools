package apollo

import (
	"reflect"
	"strings"
	"sync"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/sirupsen/logrus"
)

// onChangeFunc is the callback invoked when a watched ",dynamic" key changes.
// The callback receives the new raw string value and is responsible for
// parsing it and updating its own state (typically a Dynamic[T].store).
type onChangeFunc func(newVal string)

type apolloListener struct {
	watchKeys map[string]onChangeFunc
	lock      sync.RWMutex
}

func newApolloListener() *apolloListener {
	return &apolloListener{
		watchKeys: make(map[string]onChangeFunc),
	}
}

// TryWatchKey registers an OnChange callback for a ",dynamic" key. If the key
// does not carry the ",dynamic" suffix it is a no-op (returns false): non-
// dynamic keys are loaded once and never hot-reloaded. The callback receives
// the new raw string value on each apollo change event.
//
// Registering the same key twice (e.g. with different generic type T) would
// silently overwrite the first callback, so the first Dynamic[T] would never
// receive updates. To prevent this, a duplicate registration is rejected
// (returns false).
func (a *apolloListener) TryWatchKey(key string, onChange onChangeFunc) bool {
	if !strings.HasSuffix(key, ",dynamic") {
		return false
	}
	trimmed := strings.TrimSuffix(key, ",dynamic")

	a.lock.Lock()
	defer a.lock.Unlock()

	if _, exists := a.watchKeys[trimmed]; exists {
		logrus.Warnf("apollo key %q is already registered for OnChange; duplicate registration rejected", trimmed)
		return false
	}

	a.watchKeys[trimmed] = onChange
	return true
}

func (a *apolloListener) OnNewestChange(event *storage.FullChangeEvent) {
	// empty
}

func (a *apolloListener) OnChange(event *storage.ChangeEvent) {
	for key, change := range event.Changes {
		a.lock.RLock()
		fn, ok := a.watchKeys[key]
		a.lock.RUnlock()
		if !ok {
			continue
		}

		newValue, ok := change.NewValue.(string)
		if !ok {
			logrus.Errorf("new value is not string type, t: %v", reflect.TypeOf(change.NewValue))
			continue
		}

		fn(newValue)
		logrus.Infof("apollo config changed, key %s", key)
	}
}
