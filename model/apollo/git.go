package apollo

type GitConfig struct {
	Providers map[string]GitItemConfig `json:"providers"`
}

type GitItemConfig struct {
	Host    string `json:"host"`
	Token   string `json:"token"`
	Enable  bool   `json:"enable"`
	ShaPath string `json:"shaPath"`
}
