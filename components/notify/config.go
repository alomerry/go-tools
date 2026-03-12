package notify

import (
  "github.com/alomerry/go-tools/static/cons/notify"
)

// Config 保存通知系统的配置
type Config struct {
	Notifiers []NotifierConfig `json:"notifiers" yaml:"notifiers"`
}

// NotifierConfig 保存单个通知器实例的配置
type NotifierConfig struct {
	Name   string                 `json:"name" yaml:"name"`     // 此实例的唯一名称
	Driver notify.NotifySenderType  `json:"driver" yaml:"driver"` // 驱动名称 (例如 "email", "slack")
	Config map[string]interface{} `json:"config" yaml:"config"` // 驱动特定配置
}
