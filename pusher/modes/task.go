package modes

import "golang.org/x/net/context"

type Task interface {
	InitConfig()
	Run(context.Context) error
	Done() bool
}
