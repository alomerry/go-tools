package apollo

type KookConfig struct {
	EncryptedKey        string            `json:"encryptedKey"`
	Token               string            `json:"token"`
	VerifyToken         string            `json:"verifyToken"`
	RootNotifyChannelId string            `json:"rootNotifyChannelId"`
	RootUserId          string            `json:"rootUserId"`
	GroupChannel        map[string]string `json:"groupChannel"`
}

func (kc *KookConfig) GetGroupChannel(group string) string {
	if len(kc.GroupChannel) == 0 {
		return "5049531327335577"
	}

	return kc.GroupChannel[group]
}
