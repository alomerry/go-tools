package kook

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alomerry/go-tools/components/kook/client"
	model2 "github.com/alomerry/go-tools/components/kook/model"
  "github.com/alomerry/go-tools/components/notify"
  notify2 "github.com/alomerry/go-tools/static/cons/notify"
  "github.com/alomerry/go-tools/utils/apollo"
  "github.com/sirupsen/logrus"
)

func init() {
	notify.Register(notify2.NotifySenderKook, &Driver{})
}

type Driver struct{}

// Open 初始化 Kook 通知器
func (d *Driver) Open() (notify.Notifier, error) {
	// 这里的 client 应该也是动态获取或者复用的
	// 但是 Kook 的 client 初始化需要 Token，如果 Token 变了，Client 也要变
	// 简单起见，我们在 Send 的时候获取最新的配置，如果 Token 变了，可能需要重新 NewClient
	// 或者每次 Send 都 NewClient（开销较大）
	// 更好的方式是维护一个 Client 缓存，根据 Token 缓存
	// 这里为了简化，先假设 Token 不频繁变更，或者每次 Send 时检查

	return &Notifier{}, nil
}

type Notifier struct {
}

// Send 发送 Kook 通知
func (n *Notifier) Send(ctx context.Context, msg *notify.Message) error {
	cfg := apollo.GetKookConfig()
	if cfg == nil {
		return fmt.Errorf("kook config is nil")
	}

	if cfg.Token == "" {
		logrus.Warn("kook token is empty, skip sending")
		return nil
	}

	// 每次发送都创建一个新的 client，或者维护一个全局单例池
	// 考虑到 Kook client 创建开销（主要是 http client），可以接受
	// 如果性能敏感，可以优化
	cli := client.NewClient(client.WithToken(cfg.Token))

	// 尝试从 extra 中获取 level
	level := "info"
	if l, ok := msg.Extra["level"].(string); ok {
		level = l
	}

	theme := "warning"
	if strings.EqualFold(level, "Error") {
		theme = "danger"
	} else if strings.EqualFold(level, "Recovery") {
		theme = "success"
	}

	title := msg.Subject
	if title == "" {
		title = "系统通知"
	}

	// 构建卡片消息
	// 使用 strings.ReplaceAll 处理换行和制表符
	content := strings.ReplaceAll(msg.Content, "\n", "\\n")
	content = strings.ReplaceAll(content, "\t", "\\t")

	now := time.Now().Format("2006-01-02 15:04:05")
	card := fmt.Sprintf(`[{"type":"card","theme":"%s","modules":[{"type":"section","text":{"type":"kmarkdown","content":"%s\n%s"}},{"type":"context","elements":[{"type":"plain-text","content":"%s"}]}]}]`,
		theme, title, content, now)

	logrus.Infof("kook notify card: %s", card)

	// 获取 targetId
	group := "info"
	if g, ok := msg.Extra[notify2.NotifyMsgExtGroup].(string); ok {
		group = g
	}
	targetId := cfg.GetGroupChannel(group)
	if targetId == "" {
		logrus.Warnf("target id not found for group: %s", group)
		// 如果没有找到 targetId，这里会报错，但我们还是尝试发一下，或者直接返回错误
	}

	_, err := cli.MessageService.Create(ctx, model2.CreateMessageRequest{
		MessageType: 10, // Card Message
		TargetId:    targetId,
		Content:     card,
	})

	if err != nil {
		return err
	}

	return nil
}

func (n *Notifier) Close() error {
	return nil
}
