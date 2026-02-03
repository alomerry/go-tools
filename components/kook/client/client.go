package client

import (
	"time"

	"github.com/alomerry/go-tools/components/http"
	"github.com/alomerry/go-tools/components/http/opts/setup"
	"github.com/alomerry/go-tools/components/kook/service"
	"github.com/emicklei/go-restful/v3"
)

// Client 主客户端结构
type Client struct {
	httpCli http.Client
	baseURL string

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

	// 配置 resty 客户端
	c := &Client{
		baseURL: config.BaseURL,
		config:  config,
	}

	c.setupRestyClient()
	c.initServices()

	return c
}

// setupRestyClient 配置 resty 客户端
func (c *Client) setupRestyClient() {
	var opts []setup.Opt

	opts = append(
		opts,
		setup.Debug(c.config.Debug),
		setup.WithAccept(restful.MIME_JSON),
		setup.WithResponseMiddleware(c.processResponse),
		setup.WithBaseURL(c.baseURL),
		setup.WithContentType(restful.MIME_JSON),
		setup.WithTimeout(c.config.Timeout),
		setup.WithRetryCount(c.config.RetryCount),
	)

	// 认证头
	if c.config.AccessToken != "" {
		opts = append(opts, setup.WithAuthSchema("Bot"), setup.WithAuthToken(c.config.AccessToken))
	}

	c.httpCli = http.NewHttpClient(opts...)
}

// initServices 初始化服务
func (c *Client) initServices() {
	c.ChannelService = service.NewChannelService(c.httpCli)
	c.MessageService = service.NewMessageService(c.httpCli)
}
