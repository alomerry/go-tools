package ssh

import (
	"context"
)

type Client interface {
	Connect() error
	Close() error
	Session(ctx context.Context) (Session, error)
	Config() Config
}

type Session interface {
	Run(ctx context.Context, cmd string) (string, error)
	Close() error
}

type Config interface {
	Host() string
}
