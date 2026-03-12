package notify_test

import (
	"context"
	"testing"
  
  "github.com/alomerry/go-tools/components/apollo"
  "github.com/alomerry/go-tools/components/notify"
	_ "github.com/alomerry/go-tools/components/notify/drivers/bark"    // 注册 bark 驱动
	_ "github.com/alomerry/go-tools/components/notify/drivers/console" // 注册 console 驱动
	_ "github.com/alomerry/go-tools/components/notify/drivers/kook"    // 注册 kook 驱动
  notify2 "github.com/alomerry/go-tools/static/cons/notify"
)

func TestManager(t *testing.T) {
  apollo.Init("homelab", "backend")
	// 1. 创建管理器
	mgr := notify.NewManager()

	// 2. 初始化驱动
	err := mgr.InitDrivers()
	if err != nil {
		t.Fatalf("InitDrivers 失败: %v", err)
	}

	// 3. 发送消息
	err = mgr.Send(context.Background(), notify2.NotifySenderConsole,
		notify.WithSubject("测试主题"),
		notify.WithContent("测试内容"),
	)
	if err != nil {
		t.Errorf("Send 失败: %v", err)
	}

	// 4. 广播消息
	errs := mgr.Broadcast(context.Background(),
		notify.WithSubject("测试主题"),
		notify.WithContent("测试内容"),
	)
	if len(errs) > 0 {
		t.Errorf("Broadcast 失败: %v", errs)
	}

	// 5. 关闭
	err = mgr.Close()
	if err != nil {
		t.Errorf("Close 失败: %v", err)
	}
}

func TestDriverRegistration(t *testing.T) {
	// 1. 创建管理器
	mgr := notify.NewManager()

	// 2. 初始化驱动
	err := mgr.InitDrivers()
	if err != nil {
		t.Fatalf("InitDrivers 失败: %v", err)
	}

	// 3. 验证驱动是否存在
	if _, ok := mgr.Get(notify2.NotifySenderBark); !ok {
		t.Errorf("Bark driver not registered")
	}
	if _, ok := mgr.Get(notify2.NotifySenderKook); !ok {
		t.Errorf("Kook driver not registered")
	}
}
