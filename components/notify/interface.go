// Package notify 提供了一个通用的、可扩展的通知组件。
//
// 它定义了统一的 Notifier 和 Driver 接口，支持通过插件机制注册不同的通知驱动（如 Console, Bark, Kook 等）。
// 该组件支持配置化管理，允许在运行时通过配置文件加载和管理多个通知实例。
package notify

import (
	"context"
  "fmt"
)

// Message 定义了通知消息的结构
type Message struct {
	Subject     string                 `json:"subject"`
	Content     string                 `json:"content"`
	Attachments []string               `json:"attachments"`
	Extra       map[string]interface{} `json:"extra"`
}

// Notifier 是通知驱动必须实现的接口
type Notifier interface {
	// Send 发送通知
	Send(ctx context.Context, msg *Message) error
	// Close 关闭通知器并释放资源
	Close() error
}

// Driver 是通知器驱动的接口
type Driver interface {
	// Open 创建一个新的 Notifier 实例，配置将由实现自行从 Apollo 获取
	Open() (Notifier, error)
}

type NotifierWrapper struct {
	notifier Notifier
}

func (n NotifierWrapper) Close() error {
	return n.notifier.Close()
}

func (n NotifierWrapper) Send(ctx context.Context, opts ...Option) error {
  if n.notifier == nil {
    return fmt.Errorf("notifier instance not found")
  }
  
  msg := &Message{}
	for _, opt := range opts {
		opt.apply(msg)
	}

	return n.notifier.Send(ctx, msg)
}

func NewNotifierWrapper(notifier Notifier) NotifierWrapper {
	return NotifierWrapper{notifier: notifier}
}
