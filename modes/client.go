package modes

import (
	"errors"
	"golang.org/x/net/context"
	"sync"
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
	var (
		err     error
		errPipe = make(chan error, len(c.config.tasks))
		errDone = make(chan struct{})
		wg      = sync.WaitGroup{}
	)

	for i := range c.config.tasks {
		task := c.config.tasks[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			innerErr := task.Run(ctx)
			if innerErr != nil {
				errPipe <- innerErr
			}
		}()
	}

	go func() {
		// 合并任务错误
		for innerErr := range errPipe {
			if innerErr != nil {
				err = errors.Join(err, innerErr)
			}
		}
		errDone <- struct{}{}
	}()

	wg.Wait()

	close(errPipe)

	// 错误处理完毕后结束
	select {
	case <-errDone:
	}

	return err
}
