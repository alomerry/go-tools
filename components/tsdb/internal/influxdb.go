package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/alomerry/go-tools/static/env/influxdb"
	tsdb_err "github.com/alomerry/go-tools/static/errors/tsdb"
	"github.com/influxdata/influxdb-client-go/v2"
	"time"
)

type defaultCat struct {
	*def.Meta
	client influxdb2.Client
}

func NewDefaultCat(ctx context.Context, options ...def.Option) (*defaultCat, error) {
	cat := &defaultCat{
		Meta: &def.Meta{},
	}

	for _, opt := range options {
		opt(cat.Meta)
	}

	cat.init()

	if err := cat.validate(); err != nil {
		return nil, err
	}

	cat.client = influxdb2.NewClient(cat.Endpoint, influxdb.GetToken())

	running, err := cat.client.Ping(ctx)
	if err != nil || !running {
		return nil, errors.Join(tsdb_err.ErrUnhealthy, errors.New(fmt.Sprintf("running: %v, err: %v", running, err)))
	}

	return cat, nil
}

func (d *defaultCat) init() {
}

func (d *defaultCat) validate() error {
	if d.Endpoint == "" {
		return tsdb_err.ErrEmptyEndpoint
	}

	if d.Org == "" {
		return tsdb_err.ErrEmptyOrg
	}

	return nil
}

func (d *defaultCat) LogPoint(measurement string, tags map[string]string, fields map[string]any) error {
	return d.LogPointWithTime(measurement, tags, fields, time.Now())
}

func (d *defaultCat) LogPointWithTime(measurement string, tags map[string]string, fields map[string]any, date time.Time) error {
	bucket := "tmp" // 层级
	writeAPI := d.client.WriteAPIBlocking(d.Org, bucket)
	p := influxdb2.NewPoint(measurement, tags, fields, date)
	err := writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		return err
	}
	return nil
}
