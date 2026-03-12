package apollo

type KafkaConfig struct {
	Brokers     []string `json:"brokers"`
	MetricTopic string   `json:"metricTopic"`
	UserName    string   `json:"userName"`
	Password    string   `json:"password"`
}
