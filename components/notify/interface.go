// Package notify 提供了一个通用的、可扩展的通知组件。
//
// 它定义了统一的 Notifier 和 Driver 接口，支持通过插件机制注册不同的通知驱动（如 Console, Bark, Kook 等）。
// 该组件支持配置化管理，允许在运行时通过配置文件加载和管理多个通知实例。
package notify

import (
	"context"
	"fmt"
)

type Button struct {
	Text  string `json:"text"`
	Theme string `json:"theme,omitempty"` // primary, success, danger, warning, info, secondary
	Value string `json:"value"`
}

// RawSection is a pre-rendered kmarkdown section carried verbatim on a
// notify.Message. It lets a caller bypass the driver's whole-content
// EscapeKMarkdown for messages that need fine-grained control over their
// kmarkdown (e.g. an aggregated multi-column table whose header is bold and
// whose cells are individually escaped).
//
// When Cols > 0 the driver renders the section as a paragraph structure
// element (a cols-column grid of Fields); Cols 1..3 renders a paragraph grid.
// The Fields values are the CALLER's responsibility to escape — the driver
// must NOT re-escape them.
//
// Currently only the Kook driver renders RawSections; other drivers fall back
// to Subject/Content.
type RawSection struct {
	// Cols is the column count for a paragraph section. 1..3 renders a
	// paragraph grid.
	Cols int `json:"cols,omitempty"`
	// Fields are the kmarkdown cell contents of a paragraph section, one entry
	// per cell (the CALLER must escape each value). The cells are laid out
	// left-to-right, wrapping into rows of Cols.
	Fields []string `json:"fields,omitempty"`
}

// Message 定义了通知消息的结构
type Message struct {
	Subject     string                 `json:"subject"`
	Content     string                 `json:"content"`
	Attachments []string               `json:"attachments"`
	Extra       map[string]interface{} `json:"extra"`
	Buttons     []Button               `json:"buttons,omitempty"`
	// RawSections carries pre-rendered kmarkdown sections that the driver must
	// emit verbatim (no whole-content EscapeKMarkdown). When non-empty, a driver
	// that supports raw sections renders them instead of the escaped
	// Subject/Content pair, giving the caller full control over the kmarkdown
	// layout (e.g. a paragraph table). Drivers without raw-section support fall
	// back to the escaped Subject/Content path. The Subject is still used as the
	// card title/header.
	RawSections []RawSection `json:"rawSections,omitempty"`
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
