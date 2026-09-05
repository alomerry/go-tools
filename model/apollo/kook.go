package apollo

type KookConfig struct {
	EncryptedKey        string            `json:"encryptedKey"`
	Token               string            `json:"token"`
	VerifyToken         string            `json:"verifyToken"`
	IpBanToken          string            `json:"ipBanToken"`
	RootNotifyChannelId string            `json:"rootNotifyChannelId"`
	RootUserId          string            `json:"rootUserId"`
	GroupChannel        map[string]string `json:"groupChannel"`
	// Timezone 卡片时间戳渲染所用 IANA 时区名（如 "Asia/Shanghai"），
	// 为空时使用服务器本地时区
	Timezone string `json:"timezone"`
}

func (kc *KookConfig) GetGroupChannel(group string) string {
	if len(kc.GroupChannel) == 0 {
		return "5049531327335577"
	}

	return kc.GroupChannel[group]
}
