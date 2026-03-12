package apollo

type InfluxDbConfig struct {
	Org      string `json:"org"`
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}
