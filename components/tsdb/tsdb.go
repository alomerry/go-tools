package tsdb

import (
	"context"
	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/alomerry/go-tools/components/tsdb/internal"
)

func NewMetric(ctx context.Context, options ...def.Option) def.Metric {
	cat, err := internal.NewDefaultCat(ctx, options...)
	if err != nil {
		panic(err)
	}
	return cat
}
