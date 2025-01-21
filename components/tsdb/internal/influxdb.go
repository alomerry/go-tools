package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/alomerry/go-tools/static/env/influxdb"
	tsdb_err "github.com/alomerry/go-tools/static/errors/tsdb"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
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
	return cat, cat.Ping(context.Background())
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

func (d *defaultCat) LogPoint(bucket, measurement string, tags map[string]string, fields map[string]any) error {
	return d.LogPointWithTime(bucket, measurement, tags, fields, time.Now())
}

func (d *defaultCat) LogPointWithTime(bucket, measurement string, tags map[string]string, fields map[string]any, date time.Time) error {
	if len(bucket) == 0 {
		return tsdb_err.ErrEmptyBucket
	}
	writeAPI := d.client.WriteAPIBlocking(d.Org, bucket)
	p := influxdb2.NewPoint(measurement, tags, fields, date)
	err := writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		return err
	}
	return nil
}

func (d *defaultCat) Ping(ctx context.Context) error {
	running, err := d.client.Ping(ctx)
	if err != nil || !running {

		return errors.Join(tsdb_err.ErrUnhealthy, fmt.Errorf("running: %v, err: %v", running, err))
	}

	return nil
}
