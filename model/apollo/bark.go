package apollo

type BarkConfig struct {
	DeviceIds []string `json:"deviceIds"`

	OnlyMsgPath     string `json:"onlyMsgPath"`
	TitleAndMsgPath string `json:"titleAndMsgPath"`

	AlarmParamMap map[string]string `json:"alarmParamMap"`
	IconUrlPath   map[string]string `json:"iconUrlPath"`
}
