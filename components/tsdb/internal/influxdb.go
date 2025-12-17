package internal

/*type defaultCat struct {
	*def.Meta
	client    influxdb2.Client
	writeAPIs map[string]api.WriteAPIBlocking // cache for write APIs
	asyncAPIs map[string]api.WriteAPI         // cache for async write APIs
}

func NewDefaultCat(ctx context.Context, options ...def.Option) (def.Metric, error) {
	cat := &defaultCat{
		Meta:      &def.Meta{},
		writeAPIs: make(map[string]api.WriteAPIBlocking),
		asyncAPIs: make(map[string]api.WriteAPI),
	}

	for _, opt := range options {
		opt(cat.Meta)
	}

	cat.init()

	if err := cat.validate(); err != nil {
		return nil, err
	}

	// Use token from options or fallback to environment variable
	token := cat.Token
	if token == "" {
		token = influxdb.GetToken()
	}

	cat.client = influxdb2.NewClient(cat.Endpoint, token)
	return cat, cat.Ping(ctx)
}

func (d *defaultCat) init() {
	// Set defaults if not provided
	if d.Org == "" {
		d.Org = influxdb.GetOrg()
	}
	if d.Endpoint == "" {
		d.Endpoint = influxdb.GetEndpoint()
	}
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

// getWriteAPI returns a cached WriteAPIBlocking for the given bucket
func (d *defaultCat) getWriteAPI(bucket string) api.WriteAPIBlocking {
	if api, ok := d.writeAPIs[bucket]; ok {
		return api
	}
	api := d.client.WriteAPIBlocking(d.Org, bucket)
	d.writeAPIs[bucket] = api
	return api
}

// LogPoint writes a single point to the specified bucket
func (d *defaultCat) LogPoint(bucket, measurement string, tags map[string]string, fields map[string]any) error {
	return d.LogPointWithTime(bucket, measurement, tags, fields, time.Now())
}

// LogPointWithTime writes a single point with a specific timestamp
func (d *defaultCat) LogPointWithTime(bucket, measurement string, tags map[string]string, fields map[string]any, date time.Time) error {
	if len(bucket) == 0 {
		return tsdb_err.ErrEmptyBucket
	}
	writeAPI := d.getWriteAPI(bucket)
	p := influxdb2.NewPoint(measurement, tags, fields, date)
	return writeAPI.WritePoint(context.Background(), p)
}

// LogPoints writes multiple points in a batch (more efficient)
func (d *defaultCat) LogPoints(bucket string, points []def.Point) error {
	if len(bucket) == 0 {
		return tsdb_err.ErrEmptyBucket
	}
	if len(points) == 0 {
		return nil
	}

	writeAPI := d.getWriteAPI(bucket)
	ctx := context.Background()

	// Convert points to InfluxDB points and write them
	for _, p := range points {
		influxPoint := influxdb2.NewPoint(p.Measurement, p.Tags, p.Fields, p.Time)
		if err := writeAPI.WritePoint(ctx, influxPoint); err != nil {
			return fmt.Errorf("failed to write point %s: %w", p.Measurement, err)
		}
	}

	return nil
}

// LogPointToDefault writes to the default bucket if configured
func (d *defaultCat) LogPointToDefault(measurement string, tags map[string]string, fields map[string]any) error {
	if d.Bucket == "" {
		return tsdb_err.ErrEmptyBucket
	}
	return d.LogPoint(d.Bucket, measurement, tags, fields)
}

// Counter increments a counter metric
func (d *defaultCat) Counter(bucket, measurement string, tags map[string]string, value float64) error {
	fields := map[string]any{
		"value": value,
		"type":  "counter",
	}
	return d.LogPoint(bucket, measurement, tags, fields)
}

// Gauge sets a gauge metric value
func (d *defaultCat) Gauge(bucket, measurement string, tags map[string]string, value float64) error {
	fields := map[string]any{
		"value": value,
		"type":  "gauge",
	}
	return d.LogPoint(bucket, measurement, tags, fields)
}

// Histogram records a histogram value
func (d *defaultCat) Histogram(bucket, measurement string, tags map[string]string, value float64) error {
	fields := map[string]any{
		"value": value,
		"type":  "histogram",
	}
	return d.LogPoint(bucket, measurement, tags, fields)
}

// Summary records a summary value
func (d *defaultCat) Summary(bucket, measurement string, tags map[string]string, value float64) error {
	fields := map[string]any{
		"value": value,
		"type":  "summary",
	}
	return d.LogPoint(bucket, measurement, tags, fields)
}

// Ping checks if the InfluxDB connection is healthy
func (d *defaultCat) Ping(ctx context.Context) error {
	running, err := d.client.Ping(ctx)
	if err != nil || !running {
		return errors.Join(tsdb_err.ErrUnhealthy, fmt.Errorf("running: %v, err: %v", running, err))
	}
	return nil
}

// Close closes the client connection and releases resources
func (d *defaultCat) Close() error {
	if d.client != nil {
		d.client.Close()
	}
	return nil
}
*/
