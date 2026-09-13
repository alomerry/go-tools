package def

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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

	Start    *time.Time
	End      *time.Time
	Query    *string
	Interval string
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
|> aggregateWindow(every: %s, fn: mean, createEmpty: false)
`,
		fluxEscape(t.Bucket),
		start, end,
		fluxEscape(t.Measurement),
		t.getTags(),
		t.getFields(),
		t.getGroup(),
		t.getInterval(),
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

	// Interval 以字面量插入查询，非法值（如含引号/表达式）属查询注入，直接拒绝
	if t.Interval != "" && !fluxDurationRe.MatchString(t.Interval) {
		return fmt.Errorf("invalid flux duration %q", t.Interval)
	}

	return nil
}

// fluxDurationRe 校验 Flux duration 字面量（如 -1m / 5s / 30m / 24h），
// Interval 等拼接进查询的字面量必须匹配
var fluxDurationRe = regexp.MustCompile(`^-?\d+(ns|us|µs|ms|s|m|h|d|w|y)$`)

// fluxEscape 转义 Flux 双引号字符串字面量中的特殊字符。Bucket/Measurement/
// tag 等外部输入直接 Sprintf 进查询时，值内含 `"` 即可逃逸出字符串并拼接
// 任意 Flux 表达式（等价 SQL 注入），此处统一转义。
func fluxEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// getTimeRange returns the Flux range literals for start and stop.
// Explicit times are converted to UTC and rendered unquoted with a Z suffix
// (e.g. 2026-08-16T03:34:27Z); Flux keywords now() and relative durations
// like -1m stay as-is.
func (t *TsdbQueryOptions) getTimeRange() (string, string) {
	var (
		start, end string
	)
	if t.End == nil {
		end = "now()"
	} else {
		end = t.End.UTC().Format(time.RFC3339) // 产出 2026-08-16T03:34:27Z
	}

	if t.Start == nil {
		start = "-1m"
	} else {
		start = t.Start.UTC().Format(time.RFC3339) // 产出 2026-08-15T03:34:27Z
	}

	return start, end
}

func (t *TsdbQueryOptions) getInterval() string {
	if t.Interval != "" {
		return t.Interval
	}

	if t.Start != nil && t.End != nil {
		duration := t.End.Sub(*t.Start)
		switch {
		case duration >= 24*time.Hour:
			return "30m"
		case duration >= 6*time.Hour:
			return "5m"
		case duration >= 1*time.Hour:
			return "1m"
		}
	}

	return "5s"
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
				items = append(items, fmt.Sprintf(`r["%s"] == "%s"`, fluxEscape(k), fluxEscape(v)))
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
		// 无 tag 分组时仍按 _field 分组，保留字段身份，使 record.Field() 不为空，
		// series 名恢复为真实字段名；与非空 case（Groups + "_field"）规则统一。
		return `|> group(columns: ["_field"])`
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

func WithInterval(interval string) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		opts.Interval = interval
	}
}

type TsdbReader interface {
	Query(ctx context.Context, opts ...func(*TsdbQueryOptions)) ([]*Series, error)
	Close() error
}
