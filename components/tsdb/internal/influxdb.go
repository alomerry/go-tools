package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/alomerry/go-tools/static/errors/tsdb"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// single org influx db client

type influxdbClient struct {
	token     string
	org       string
	endpoint  string
	bucket    string
	client    influxdb2.Client
	writeAPIs map[string]api.WriteAPIBlocking // cache for write APIs
	asyncAPIs map[string]api.WriteAPI         // cache for async write APIs
}

func NewInfluxdbClient(ctx context.Context, org, endpoint, bucket, token string) (*influxdbClient, error) {
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
	if wa, ok := d.writeAPIs[bucket]; ok {
		return wa
	}
	wa := d.client.WriteAPIBlocking(d.org, bucket)
	d.writeAPIs[bucket] = wa
	return wa
}

func (d *influxdbClient) getQueryAPI() api.QueryAPI {
	return d.client.QueryAPI(d.org)
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

func (d *influxdbClient) Query(ctx context.Context, opts ...func(*def.TsdbQueryOptions)) ([]*def.Series, error) {
	options := new(def.TsdbQueryOptions)
	for _, opt := range opts {
		opt(options)
	}

	query, err := options.GetQuery()
	if err != nil {
		return nil, err
	}

	results, err := d.getQueryAPI().Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var (
		series = make(map[string]*def.Series)
		res    []*def.Series
	)

	for results.Next() {
		record := results.Record()

		keys := make([]string, 0, len(options.Groups)*len(options.Fields))
		for _, group := range options.Groups {
			keys = append(keys, cast.ToString(record.Values()[group]))
		}

		field := record.Field()
		if len(record.Field()) == 0 {
			field = "deafult_count"
		}

		keys = append(keys, field)
		key := strings.Join(keys, "-")
		if _, ok := series[key]; !ok {
			series[key] = &def.Series{
				Name:    key,
				Tags:    make(map[string]string),
				Columns: append([]string{"time", field}),
			}

			for i, group := range options.Groups {
				series[key].Tags[group] = keys[i]
			}
		}

		var item = []any{
			record.Time().Unix(),
			record.Values()["_value"],
		}

		series[key].Values = append(series[key].Values, item)
		// fmt.Println(record)
	}
	if err = results.Err(); err != nil {
		return nil, err
	}

	res = make([]*def.Series, 0, len(series))
	for k := range series {
		res = append(res, series[k])
	}

	return res, nil
}

func (d *influxdbClient) readBy(ctx context.Context, bucket string) {
	queryAPI := d.getQueryAPI()
	query := `from(bucket: "my-bucket")
            |> range(start: -10m)
            |> filter(fn: (r) => r._measurement == "measurement1")`
	results, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		logrus.Fatal(err)
	}
	for results.Next() {
		fmt.Println(results.Record())
	}
	if err := results.Err(); err != nil {
		logrus.Fatal(err)
	}
}
