package kook

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alomerry/go-tools/components/ext"
	"github.com/alomerry/go-tools/components/kook/client"
	model2 "github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/components/notify"
	notify2 "github.com/alomerry/go-tools/static/cons/notify"
	strutil "github.com/alomerry/go-tools/utils/string"
	time2 "github.com/alomerry/go-tools/utils/time"
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
	cfg := ext.Apollo().KookCfg()
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

	theme := model2.CardThemeWarning
	if strings.EqualFold(level, "Error") {
		theme = model2.CardThemeDanger
	} else if strings.EqualFold(level, "Recovery") {
		theme = model2.CardThemeSuccess
	}

	title := msg.Subject
	if title == "" {
		title = "系统通知"
	}
	escTitle := strutil.EscapeKMarkdown(title)
	escContent := strutil.EscapeKMarkdown(msg.Content)

	// Render the timestamp in fixed UTC+8: scratch/busybox images have no
	// zoneinfo, so time.Local falls back to UTC and a bare Format would show a
	// time 8 hours off. Asia/Shanghai has no DST, so FixedZone(+08:00) is exact
	// and avoids the tzdata dependency of utils/time.DefaultTimeZone.
	now := time.Now().In(time.FixedZone("CST", 8*3600)).Format(time2.Readable)

	builder := model2.NewCardBuilder().
		Theme(theme).
		AddSectionKmarkdown(fmt.Sprintf("%s\n%s", escTitle, escContent)).
		AddContextKmarkdown(now)

	if len(msg.Buttons) > 0 {
		buttons := make([]model2.ElementButton, 0, len(msg.Buttons))
		for _, btn := range msg.Buttons {
			btnTheme := model2.CardThemePrimary
			if btn.Theme != "" {
				btnTheme = model2.CardTheme(btn.Theme)
			}
			buttons = append(buttons, model2.ElementButton{
				Type:  model2.ElementTypeButton,
				Theme: btnTheme,
				Value: btn.Value,
				Click: "return-val",
				Text: model2.ElementText{
					Type:    model2.ElementTypePlainText,
					Content: btn.Text,
				},
			})
		}
		builder.AddActionGroup(buttons)
	}

	card := model2.NewCardMessageBuilder().AddCard(builder.Build()).Build()

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
