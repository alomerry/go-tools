package es

import (
	"github.com/alomerry/go-tools/static/env"
	str_utils "github.com/alomerry/go-tools/utils/string"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sirupsen/logrus"
)

type Client struct {
	es *elasticsearch.TypedClient

	endpoint string
	apiKey   string
}

type Option func(*Client)

func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

func NewClient(opts ...Option) *Client {
	client := &Client{}
	for _, opt := range opts {
		opt(client)
	}
	cfg := elasticsearch.Config{
		Addresses: []string{
			str_utils.FirstNotBlank(client.endpoint, env.GetElasticSearchEndpoint()),
		},
		APIKey: str_utils.FirstNotBlank(client.apiKey, env.GetElasticSearchAK()),
	}
	tc, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		logrus.Panicf("failed to create elasticsearch client: %v", err)
	}

	client.es = tc
	return client
}

func (c *Client) GetEs() *elasticsearch.TypedClient {
	return c.es
}
