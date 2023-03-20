package modes

import "golang.org/x/net/context"

type Task interface {
	InitConfig()
	Run(ctx context.Context) error
	Done(ctx context.Context) bool
}
