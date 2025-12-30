package def

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alomerry/go-tools/static/cons/tsdb"
	"github.com/alomerry/go-tools/static/env"
)

type tagOpKey struct {
	Key string
	Op  tsdb.TagOp
}

type TsdbQueryOptions struct {
	Bucket      string
	Measurement string
	Fields      []string
	Groups      []string

	tags map[tagOpKey][]string

	Start *time.Time
	End   *time.Time
	Query *string
}

func (t *TsdbQueryOptions) Apply(opts ...func(*TsdbQueryOptions)) {
	for _, opt := range opts {
		opt(t)
	}
}

func (t *TsdbQueryOptions) GetQuery() (string, error) {
	if err := t.validate(); err != nil {
		return "", err
	}

	if t.Query != nil {
		return *t.Query, nil
	}

	start, end := t.getTimeRange()

	query := fmt.Sprintf(`from(bucket: "%s")
|> range(start: %s, stop: %s)
|> filter(fn: (r) => r._measurement == "%s")
%s
%s
%s
|> aggregateWindow(every: 5s, fn: mean)
`,
		t.Bucket,
		start, end,
		t.Measurement,
		t.getTags(),
		t.getFields(),
		t.getGroup(),
	)

	if env.Local() {
		fmt.Println(query)
	}

	return query, nil
}

func (t *TsdbQueryOptions) validate() error {
	if t.Bucket == "" || t.Measurement == "" {
		return fmt.Errorf("bucket and measurement are required")
	}

	if t.Start != nil && t.End != nil && t.Start.After(*t.End) {
		return fmt.Errorf("start time must be before end time")
	}

	return nil
}

func (t *TsdbQueryOptions) getTimeRange() (string, string) {
	var (
		start, end string
	)
	if t.End == nil {
		end = "now()"
	} else {
		end = t.End.Format(time.RFC3339)
	}

	if t.Start == nil {
		start = "-1m"
	} else {
		start = t.Start.Format(time.RFC3339)
	}

	return start, end
}

func (t *TsdbQueryOptions) getTags() string {
	if len(t.tags) == 0 {
		return ""
	}

	var tags []string
	for ko, vs := range t.tags {
		k := ko.Key
		op := ko.Op

		var items []string
		switch op {
		case tsdb.OpEqual:
			for _, v := range vs {
				items = append(items, fmt.Sprintf(`r["%s"] == "%s"`, k, v))
			}
		}
		tags = append(tags, fmt.Sprintf("|> filter(fn: (r) => %s)", strings.Join(items, " or ")))
	}

	return strings.Join(tags, "\n")
}

func (t *TsdbQueryOptions) getFields() string {
	var (
		fields string
	)
	if len(t.Fields) == 0 {
		return ""
	}

	val, _ := json.Marshal(t.Fields)

	fields = fmt.Sprintf("|> filter(fn: (r) => contains(value: r._field, set: %s))", val)

	return fields
}

func (t *TsdbQueryOptions) getGroup() string {
	var (
		group string
	)
	if len(t.Groups) == 0 {
		return `|> group(columns: [])`
	}

	val, _ := json.Marshal(append(t.Groups, "_field"))

	group = fmt.Sprintf("|> group(columns: %s)", val)

	return group
}

func WithTag(key string, op tsdb.TagOp, values ...string) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		if opts.tags == nil {
			opts.tags = make(map[tagOpKey][]string)
		}

		if _, ok := opts.tags[tagOpKey{key, op}]; !ok {
			opts.tags[tagOpKey{key, op}] = values
		} else {
			opts.tags[tagOpKey{key, op}] = append(opts.tags[tagOpKey{key, op}], values...)
		}
	}
}

func WithBucket(bucket string) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		opts.Bucket = bucket
	}
}

func WithMeasurement(measurement string) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		opts.Measurement = measurement
	}
}

func WithGroup(group ...string) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		opts.Groups = append(opts.Groups, group...)
	}
}

func WithFields(fields ...string) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		opts.Fields = fields
	}
}

func WithStart(start time.Time) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		opts.Start = &start
	}
}

func WithEnd(end time.Time) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		opts.End = &end
	}
}

func WithQuery(query string) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		if len(query) > 0 {
			opts.Query = &query
		}
	}
}

type TsdbReader interface {
	Query(ctx context.Context, opts ...func(*TsdbQueryOptions)) ([]*Series, error)
	Close() error
}
