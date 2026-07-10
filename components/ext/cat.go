package ext

import (
	"context"

	"github.com/alomerry/cat-go/cat"
	"github.com/alomerry/go-tools/static/cons"
	"github.com/alomerry/go-tools/static/env"
)

func init() {
	Register(cons.ExtCat, NewCatExt)
}

type catExt struct {
}

func NewCatExt() Ext {
	return catExt{}
}

func (catExt) Init(_ context.Context) error {
	domain := env.GetService()
	if len(domain) == 0 {
		domain = "cat"
	}

	if env.Debug() {
		cat.DebugOn()
	}
	cat.Init(domain)
	return nil
}
