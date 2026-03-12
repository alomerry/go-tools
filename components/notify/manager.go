package notify

import (
	"context"
  "fmt"
	"sync"
  
  "github.com/alomerry/cat-go/cat"
  "github.com/alomerry/go-tools/static/cons/notify"
)

var (
  manager *Manager
  managerOnce sync.Once
)

// Manager 管理多个通知器实例，提供统一的消息发送和广播接口。
// 它负责加载配置、实例化驱动以及管理通知器的生命周期。
type Manager struct {
	notifiers map[notify.NotifySenderType]NotifierWrapper
	mu        sync.RWMutex
}

// NewManager 创建一个新的 Manager
func NewManager() *Manager {
  managerOnce.Do(func() {
    manager = &Manager{
      notifiers: make(map[notify.NotifySenderType]NotifierWrapper),
    }
  })
	return manager
}

// Register 注册一个新的通知器实例
func (m *Manager) Register(name notify.NotifySenderType, n Notifier) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifiers[name] = NewNotifierWrapper(n)
}

// Unregister 注销一个通知器实例
func (m *Manager) Unregister(name notify.NotifySenderType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.notifiers, name)
}

// Get 通过名称获取通知器实例
func (m *Manager) Get(name notify.NotifySenderType) (NotifierWrapper, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.notifiers[name]
	return n, ok
}

// InitDrivers 初始化所有已注册的驱动
func (m *Manager) InitDrivers() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, name := range Drivers() {
		driver, _ := drivers[name] // Drivers() ensures existence
		notifier, err := driver.Open()
		if err != nil {
			return fmt.Errorf("failed to open notifier %s: %w", name, err)
		}
		m.notifiers[name] = NewNotifierWrapper(notifier)
	}
	return nil
}

func (m *Manager) WithNotifier(ctx context.Context, senderType notify.NotifySenderType) NotifierWrapper {
	m.mu.RLock()
	n, ok := m.notifiers[senderType]
	m.mu.RUnlock()
	if !ok {
    cat.LogError(ctx, fmt.Errorf("%s notifier not found", senderType))
    return NotifierWrapper{}
	}

	return n
}

// Send 使用指定的通知器发送消息
func (m *Manager) Send(ctx context.Context, name notify.NotifySenderType, opts ...Option) error {
	m.mu.RLock()
	n, ok := m.notifiers[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("notifier instance %s not found", name)
	}

	return n.Send(ctx, opts...)
}

// Broadcast 向所有注册的通知器发送消息
// 它返回一个以通知器名称为键的错误映射，如果全部成功则返回 nil
func (m *Manager) Broadcast(ctx context.Context, opts ...Option) map[notify.NotifySenderType]error {
	m.mu.RLock()
	// 复制通知器以避免在发送期间持有锁
	notifiers := make(map[notify.NotifySenderType]NotifierWrapper, len(m.notifiers))
	for k, v := range m.notifiers {
		notifiers[k] = v
	}
	m.mu.RUnlock()

	var (
		wg     sync.WaitGroup
		errs   = make(map[notify.NotifySenderType]error)
		errsMu sync.Mutex
	)

	for name, n := range notifiers {
		wg.Add(1)
		go func(name notify.NotifySenderType, n NotifierWrapper) {
			defer wg.Done()
			if err := n.Send(ctx, opts...); err != nil {
				errsMu.Lock()
				errs[name] = err
				errsMu.Unlock()
			}
		}(name, n)
	}
	wg.Wait()

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Close 关闭所有管理的通知器
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for _, n := range m.notifiers {
		if err := n.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 清空通知器
	m.notifiers = make(map[notify.NotifySenderType]NotifierWrapper)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing notifiers: %v", errs)
	}
	return nil
}
