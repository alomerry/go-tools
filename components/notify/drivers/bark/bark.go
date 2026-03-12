package bark

import (
	"context"
	"fmt"

	"github.com/alomerry/go-tools/components/http"
	req2 "github.com/alomerry/go-tools/components/http/opts/req"
  notify2 "github.com/alomerry/go-tools/components/notify"
  "github.com/alomerry/go-tools/static/cons/notify"
  "github.com/alomerry/go-tools/utils/apollo"
  "github.com/sirupsen/logrus"
)

func init() {
  notify2.Register(notify.NotifySenderBark, &Driver{})
}

type Driver struct{}

// Open 初始化 Bark 通知器
func (d *Driver) Open() (notify2.Notifier, error) {
  // CfgManager 是全局单例，直接调用获取最新配置
  // 假设 GetBarkCfg() 总是返回可用的配置或者 nil
  // 实际使用时需要注意 GetBarkCfg 可能返回 nil 的情况
	return &Notifier{
	}, nil
}

type Notifier struct{}

// Send 发送 Bark 通知
func (n *Notifier) Send(ctx context.Context, msg *notify2.Message) error {
 
	cfg := apollo.GetBarkCfg()
	if cfg == nil {
		return fmt.Errorf("bark config is nil")
	}

	// 尝试从 extra 中获取 level，如果没有则默认为 info
	level := "info"
	if l, ok := msg.Extra["level"].(string); ok {
		level = l
	}

	for _, deviceId := range cfg.DeviceIds {
		params := map[string]string{
			"content":  msg.Content,
			"level":    level,
			"deviceId": deviceId,
		}

		// 根据是否有标题决定使用哪个路径
		queryUrl := cfg.OnlyMsgPath
		if msg.Subject != "" {
			queryUrl = cfg.TitleAndMsgPath
			params["title"] = msg.Subject
		}

		logrus.Infof("notify content: %s", params["content"])

		var reqOpts []req2.Opt
		reqOpts = append(reqOpts, req2.WithQueryParam(params))
		reqOpts = append(reqOpts, req2.WithPathParam(params))

		// 设置图标（如果有配置）
		if iconUrl, ok := cfg.IconUrlPath[level]; ok {
			reqOpts = append(reqOpts, req2.WithQueryParam(map[string]string{"icon": iconUrl}))
		}

		_, err := http.GetClient().Get(ctx, queryUrl, reqOpts...)
		if err != nil {
			logrus.Errorf("notify device error: %v", err)
			continue // 继续发送到其他设备
		}
	}

	return nil
}

func (n *Notifier) Close() error {
	return nil
}
