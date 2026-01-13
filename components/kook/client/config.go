package client

import "time"

// WithToken 设置访问令牌
func WithToken(token string) Option {
	return func(c *Config) {
		c.AccessToken = token
	}
}

// WithBaseURL 设置基础 URL
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithRetryCount 设置重试次数
func WithRetryCount(count int) Option {
	return func(c *Config) {
		c.RetryCount = count
	}
}

// WithDebug 启用调试模式
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}
