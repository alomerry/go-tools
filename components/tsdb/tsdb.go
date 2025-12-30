package tsdb

import (
	"context"

	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/alomerry/go-tools/components/tsdb/internal"
)

func NewTsDbWriter(ctx context.Context, options ...func(any)) (def.TsdbWriter, error) {
	m := &meta{}
	for _, opt := range options {
		opt(m)
	}

	return internal.NewInfluxdbClient(ctx, m.org, m.endpoint, m.bucket, m.token)
}

func NewTsDbReader(ctx context.Context, options ...func(any)) (def.TsdbReader, error) {
	m := &meta{}
	for _, opt := range options {
		opt(m)
	}

	return internal.NewInfluxdbClient(ctx, m.org, m.endpoint, m.bucket, m.token)
}
