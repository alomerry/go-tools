package client

import (
	"time"

	"github.com/alomerry/go-tools/components/kook/service"
	resty2 "github.com/alomerry/go-tools/utils/resty"
	"resty.dev/v3"
)

// Client 主客户端结构
type Client struct {
	restyClient *resty.Client
	baseURL     string

	// 服务实例
	ChannelService service.ChannelService
	MessageService service.MessageService

	// 配置
	config *Config
}

// Config 客户端配置
type Config struct {
	BaseURL     string
	AccessToken string
	Timeout     time.Duration
	RetryCount  int
	Debug       bool
}

// Option 配置选项函数类型
type Option func(*Config)

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "https://www.kookapp.cn",
		Timeout:    30 * time.Second,
		RetryCount: 3,
		Debug:      false,
	}
}

// NewClient 创建新客户端
func NewClient(opts ...Option) *Client {
	// 创建配置
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// 创建 resty 客户端
	restyClient := resty.New()

	// 配置 resty 客户端
	c := &Client{
		restyClient: restyClient,
		baseURL:     config.BaseURL,
		config:      config,
	}

	c.setupRestyClient()
	c.initServices()

	return c
}

// setupRestyClient 配置 resty 客户端
func (c *Client) setupRestyClient() {
	rc := c.restyClient

	// 基础配置
	rc.SetBaseURL(c.baseURL)
	rc.SetTimeout(c.config.Timeout)
	rc.SetRetryCount(c.config.RetryCount)

	// 设置默认 Headers
	rc.SetHeader("Accept", "application/json")

	// 认证头
	if c.config.AccessToken != "" {
		rc.SetAuthScheme("Bot")
		rc.SetAuthToken(c.config.AccessToken)
		// rc.SetHeader("Authorization", fmt.Sprintf("Bot %s", c.config.AccessToken))
	}

	// 调试模式
	rc.SetDebug(c.config.Debug)

	// 设置中间件
	rc.AddRequestMiddleware(resty2.DefaultRequestMiddleware)
	rc.AddResponseMiddleware(c.processResponse)

	// 配置重试
	rc.AddRetryConditions(resty2.DefaultRetryCondition)
}

// initServices 初始化服务
func (c *Client) initServices() {
	c.ChannelService = service.NewChannelService(c.restyClient)
	c.MessageService = service.NewMessageService(c.restyClient)
}
