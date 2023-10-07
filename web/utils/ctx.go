package utils

import (
	"context"
)

type Work struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}
