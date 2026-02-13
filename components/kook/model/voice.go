package model

type JoinVoiceReq struct {
	ChannelId string `json:"channel_id"`
	Password  string `json:"password,omitempty"`
}

type JoinVoiceResp struct {
	Ip        string `json:"ip"`
	Port      string `json:"port"`
	RtcpPort  string `json:"rtcp_port"`
	RtcpMux   bool   `json:"rtcp_mux"`
	Bitrate   int    `json:"bitrate"`
	AudioSsrc string `json:"audio_ssrc"`
	AudioPt   string `json:"audio_pt"`
}

type LeaveVoiceReq struct {
	ChannelId string `json:"channel_id"`
}

type VoiceChannel struct {
	Id       string `json:"id"`
	GuildId  string `json:"guild_id"`
	ParentId string `json:"parent_id"`
	Name     string `json:"name"`
}

type VoiceListResp struct {
	Items []VoiceChannel `json:"items"`
	Meta  Meta           `json:"meta"`
}
