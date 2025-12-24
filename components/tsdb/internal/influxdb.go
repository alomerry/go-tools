package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/alomerry/go-tools/static/errors/tsdb"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type influxdbClient struct {
	token     string
	org       string
	endpoint  string
	bucket    string
	client    influxdb2.Client
	writeAPIs map[string]api.WriteAPIBlocking // cache for write APIs
	asyncAPIs map[string]api.WriteAPI         // cache for async write APIs
}

func NewInfluxdbClient(ctx context.Context, org, endpoint, bucket, token string) (def.TsdbWriter, error) {
	c := &influxdbClient{
		writeAPIs: make(map[string]api.WriteAPIBlocking),
		asyncAPIs: make(map[string]api.WriteAPI),
		token:     token,
		org:       org,
		endpoint:  endpoint,
		bucket:    bucket,
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	c.client = influxdb2.NewClient(c.endpoint, token)
	return c, c.Ping(ctx)
}

func (d *influxdbClient) validate() error {
	if d.endpoint == "" {
		return tsdb.ErrEmptyEndpoint
	}

	if d.org == "" {
		return tsdb.ErrEmptyOrg
	}

	return nil
}

func (d *influxdbClient) getWriteAPI(bucket string) api.WriteAPIBlocking {
	if api, ok := d.writeAPIs[bucket]; ok {
		return api
	}
	api := d.client.WriteAPIBlocking(d.org, bucket)
	d.writeAPIs[bucket] = api
	return api
}

func (d *influxdbClient) LogPoint(ctx context.Context, bucket, measurement string, tags map[string]string, fields map[string]any) error {
	return d.LogPointWithTime(ctx, bucket, measurement, tags, fields, time.Now())
}

func (d *influxdbClient) LogPointWithTime(ctx context.Context, bucket, measurement string, tags map[string]string, fields map[string]any, date time.Time) error {
	if len(bucket) == 0 {
		return tsdb.ErrEmptyBucket
	}
	writeAPI := d.getWriteAPI(bucket)
	p := influxdb2.NewPoint(measurement, tags, fields, date)
	return writeAPI.WritePoint(context.Background(), p)
}

func (d *influxdbClient) LogPoints(ctx context.Context, bucket string, points []def.Point) error {
	if len(bucket) == 0 {
		return tsdb.ErrEmptyBucket
	}
	if len(points) == 0 {
		return nil
	}

	writeAPI := d.getWriteAPI(bucket)

	// Convert points to InfluxDB points and write them
	for _, p := range points {
		influxPoint := influxdb2.NewPoint(p.Measurement, p.Tags, p.Fields, p.Time)
		if err := writeAPI.WritePoint(ctx, influxPoint); err != nil {
			return fmt.Errorf("failed to write point %s: %w", p.Measurement, err)
		}
	}

	return nil
}

func (d *influxdbClient) LogPointToDefault(ctx context.Context, measurement string, tags map[string]string, fields map[string]any) error {
	if d.bucket == "" {
		return tsdb.ErrEmptyBucket
	}
	return d.LogPoint(ctx, d.bucket, measurement, tags, fields)
}

func (d *influxdbClient) Counter(ctx context.Context, bucket, measurement string, tags map[string]string, value float64) error {
	fields := map[string]any{
		"value": value,
		"type":  "counter",
	}
	return d.LogPoint(ctx, bucket, measurement, tags, fields)
}

func (d *influxdbClient) Gauge(ctx context.Context, bucket, measurement string, tags map[string]string, value float64) error {
	fields := map[string]any{
		"value": value,
		"type":  "gauge",
	}
	return d.LogPoint(ctx, bucket, measurement, tags, fields)
}

func (d *influxdbClient) Histogram(ctx context.Context, bucket, measurement string, tags map[string]string, value float64) error {
	fields := map[string]any{
		"value": value,
		"type":  "histogram",
	}
	return d.LogPoint(ctx, bucket, measurement, tags, fields)
}

func (d *influxdbClient) Summary(ctx context.Context, bucket, measurement string, tags map[string]string, value float64) error {
	fields := map[string]any{
		"value": value,
		"type":  "summary",
	}
	return d.LogPoint(ctx, bucket, measurement, tags, fields)
}

func (d *influxdbClient) Ping(ctx context.Context) error {
	running, err := d.client.Ping(ctx)
	if err != nil || !running {
		return errors.Join(tsdb.ErrUnhealthy, fmt.Errorf("running: %v, err: %v", running, err))
	}
	return nil
}

func (d *influxdbClient) Close() error {
	if d.client != nil {
		d.client.Close()
	}
	return nil
}
