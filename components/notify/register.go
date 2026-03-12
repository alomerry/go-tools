package notify

import (
	"fmt"
	"sync"
  
  "github.com/alomerry/go-tools/static/cons/notify"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[notify.NotifySenderType]Driver)
)

// Register 使通知驱动通过提供的名称可用。
// 如果使用相同的名称调用两次 Register，或者驱动为 nil，它将引发 panic。
func Register(name notify.NotifySenderType, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("notify: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("notify: Register called twice for driver " + string(name))
	}
	drivers[name] = driver
}

// Drivers 返回已注册驱动名称的排序列表。
func Drivers() []notify.NotifySenderType {
	driversMu.RLock()
	defer driversMu.RUnlock()
	list := make([]notify.NotifySenderType, 0, len(drivers))
	for name := range drivers {
		list = append(list, name)
	}
	return list
}

// Open 根据指定的驱动名称打开一个通知器。
func Open(driverName notify.NotifySenderType) (Notifier, error) {
	driversMu.RLock()
	driver, ok := drivers[driverName]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("notify: unknown driver %q (forgotten import?)", driverName)
	}
	return driver.Open()
}
