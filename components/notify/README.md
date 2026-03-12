# Notify 通用通知组件

`notify` 是一个设计为抽象、低耦合、通用、高扩展且支持配置化的可热拔插通知组件。它允许通过简单的配置管理多种通知渠道（如 Console, Bark, Kook 等），并提供统一的接口进行消息发送。

## 核心设计

该组件采用了类似 `database/sql` 的驱动注册模式，实现了核心逻辑与具体实现的解耦：

1.  **抽象接口 (`interface.go`)**: 定义了 `Notifier` 和 `Driver` 接口，以及通用的 `Message` 结构体。业务层仅依赖这些接口，而不依赖具体的发送逻辑。
2.  **插件机制 (`register.go`)**: 提供了驱动注册机制。新的通知渠道（如 Slack, Email）只需实现 `Driver` 接口并在 `init()` 中调用 `Register` 即可被加载。
3.  **配置化管理 (`config.go`)**: 定义了 `Config` 和 `NotifierConfig` 结构，支持通过配置文件定义多个通知实例。例如，可以同时配置两个不同的 Kook 机器人，分别用于不同的业务场景。
4.  **生命周期管理 (`manager.go`)**: `Manager` 负责加载配置、实例化驱动、管理通知实例的生命周期，并提供了 `Send`（单发）和 `Broadcast`（广播）方法。

## 目录结构

```text
core/components/notify/
├── config.go           # 配置结构定义
├── drivers/            # 具体通知渠道实现
│   ├── bark/           # Bark 推送驱动
│   ├── console/        # [示例] 控制台输出驱动
│   └── kook/           # Kook 机器人驱动
├── interface.go        # 核心接口定义 (Notifier, Driver, Message)
├── manager.go          # 管理器实现 (加载配置, 路由消息)
├── register.go         # 驱动注册中心
└── notify_test.go      # 测试用例
```

## 快速开始

### 1. 引入依赖

在你的代码中引入 `notify` 包以及你需要使用的驱动包（使用 `_` 导入以触发注册）：

```go
import (
    "backend/core/components/notify"
    _ "backend/core/components/notify/drivers/bark"    // 注册 Bark 驱动
    _ "backend/core/components/notify/drivers/console" // 注册 Console 驱动
    _ "backend/core/components/notify/drivers/kook"    // 注册 Kook 驱动
)
```

### 2. 定义配置

配置通常来源于你的配置文件（如 YAML, JSON）。这里以 Go 结构体为例：

```go
cfg := &notify.Config{
    Notifiers: []notify.NotifierConfig{
        {
            Name:   "debug-console",
            Driver: "console",
            Config: map[string]interface{}{},
        },
        {
            Name:   "my-iphone",
            Driver: "bark",
            Config: map[string]interface{}{
                "deviceIds":       []string{"YOUR_DEVICE_KEY"},
                "onlyMsgPath":     "https://api.day.app/push/{{deviceId}}/{{content}}",
                "titleAndMsgPath": "https://api.day.app/push/{{deviceId}}/{{title}}/{{content}}",
            },
        },
        {
            Name:   "ops-group",
            Driver: "kook",
            Config: map[string]interface{}{
                "token": "YOUR_BOT_TOKEN",
                "groupChannel": map[string]string{
                    "default": "CHANNEL_ID_DEFAULT",
                    "alarm":   "CHANNEL_ID_ALARM",
                },
            },
        },
    },
}
```

### 3. 初始化与发送

```go
// 初始化管理器
mgr := notify.NewManager()
err := mgr.LoadFromConfig(cfg)
if err != nil {
    panic(err)
}

// 构造消息
msg := &notify.Message{
    Subject: "系统告警",
    Content: "检测到 CPU 使用率过高",
    Extra: map[string]interface{}{
        "level": "Error", // 影响 Bark 图标或 Kook 卡片颜色
        "group": "alarm", // 影响 Kook 发送的频道
    },
}

// 发送给特定实例
mgr.Send(ctx, "my-iphone", msg)

// 广播给所有实例
mgr.Broadcast(ctx, msg)
```

## 支持的驱动

### Console (`console`)
将通知内容直接输出到标准输出（Stdout），主要用于开发调试。

### Bark (`bark`)
向 iOS 设备发送 Bark 推送。
- **配置参数**:
  - `deviceIds`: 设备 Key 列表。
  - `onlyMsgPath`: 仅发送内容时的 API 路径模板。
  - `titleAndMsgPath`: 发送标题和内容时的 API 路径模板。
  - `iconUrlPath`: 消息级别到图标 URL 的映射（可选）。
- **Extra 参数**:
  - `level`: 消息级别，用于匹配图标。

### Kook (`kook`)
向 Kook（原开黑啦）频道发送富文本卡片消息。
- **配置参数**:
  - `token`: 机器人 Token。
  - `groupChannel`: 分组名称到频道 ID 的映射。
- **Extra 参数**:
  - `level`: 消息级别（Error/Recovery/Info），决定卡片的主题颜色（Danger/Success/Warning）。
  - `group`: 分组名称，用于路由到不同的频道。

## 如何开发新驱动

只需在 `drivers/` 下新建目录（如 `slack`），实现 `Driver` 和 `Notifier` 接口，并在 `init` 中注册：

```go
package slack

import "backend/core/components/notify"

func init() {
    notify.Register("slack", &Driver{})
}

type Driver struct{}

func (d *Driver) Open(config map[string]interface{}) (notify.Notifier, error) {
    // 解析配置并返回 Notifier 实例
    return &Notifier{}, nil
}

type Notifier struct{}

func (n *Notifier) Send(ctx context.Context, msg *notify.Message) error {
    // 实现发送逻辑
    return nil
}

func (n *Notifier) Close() error {
    return nil
}
```
