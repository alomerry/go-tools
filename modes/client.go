package modes

import (
	"errors"
	"golang.org/x/net/context"
)

var (
	IClient = &Client{}
)

type Client struct {
	config *config
}

func (c *Client) Init(ctx context.Context) *Client {
	c.config = initConfig(ctx)
	return c
}

func (c *Client) Run(ctx context.Context) error {
	var err error
	for _, task := range c.config.tasks {
		innerErr := task.Run(ctx)
		if innerErr != nil {
			err = errors.Join(err, innerErr)
		}
	}
	return err
}
