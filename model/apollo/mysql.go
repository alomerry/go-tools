package apollo

type MysqlConfig struct {
	Uri     string `json:"uri"`
	Debug   bool   `json:"debug" default:"false"`
	Verbose bool   `json:"verbose" default:"false"`
}
