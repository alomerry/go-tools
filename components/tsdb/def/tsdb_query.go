package def

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
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

	Start       *time.Time
	End         *time.Time
	Query       *string
	Interval    string
	AggregateFn tsdb.AggregateFn
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
|> aggregateWindow(every: %s, fn: %s, createEmpty: false)
`,
		fluxEscape(t.Bucket),
		start, end,
		fluxEscape(t.Measurement),
		t.getTags(),
		t.getFields(),
		t.getGroup(),
		t.getInterval(),
		t.getAggregateFn(),
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
	if t.Interval != "" && !IsValidFluxDuration(t.Interval) {
		return fmt.Errorf("invalid flux duration %q", t.Interval)
	}

	if err := t.validateAggregateFn(); err != nil {
		return err
	}

	return t.validateTags()
}

// validateAggregateFn 白名单校验聚合函数：fn 以裸字面量拼进 aggregateWindow，
// 未知值既是笔误也是注入面，一律拒绝；零值视为未指定（缺省 mean）。
func (t *TsdbQueryOptions) validateAggregateFn() error {
	// 零值视为未指定（缺省 mean），白名单外一律拒绝（fn 以裸字面量拼进查询）
	if t.AggregateFn != "" && !tsdb.IsValidAggregateFn(string(t.AggregateFn)) {
		return fmt.Errorf("unsupported aggregate fn %q", t.AggregateFn)
	}
	return nil
}

// validateTags 在拼查询前校验所有 tag 条件，任何一条不合法即拒绝生成查询
// （Query 出口报参数错误，不透传给 InfluxDB）：
//   - key 不得为空：r[""] 恒不匹配，属调用方 bug
//   - values 不得为空列表：会拼出非法 Flux `filter(fn: (r) => )`
//   - 任一 value 不得为空串：InfluxDB 空 tag value 不入索引、比较无意义，
//     与"全为空"相比按任一空串拒绝更根治（早失败，不静默丢弃条件）
//   - op 必须在已知运算符白名单内：未知 op 无拼接分支，漏检会退化成空 filter
//   - 正则 pattern 必须可通过 regexp.Compile（Flux 底层即 Go regexp，语法同源），
//     非法正则在服务端必然报错，提前拦截
func (t *TsdbQueryOptions) validateTags() error {
	for _, ko := range sortedTagKeys(t.tags) {
		if ko.Key == "" {
			return fmt.Errorf("tag key must not be empty")
		}

		values := t.tags[ko]
		if len(values) == 0 {
			return fmt.Errorf("tag %q: at least one value is required", ko.Key)
		}

		if !tsdb.IsValidTagOp(string(ko.Op)) {
			return fmt.Errorf("tag %q: unsupported tag op %q", ko.Key, ko.Op)
		}

		for _, v := range values {
			if v == "" {
				return fmt.Errorf("tag %q: value must not be empty", ko.Key)
			}
			if ko.Op == tsdb.OpRegexMatch || ko.Op == tsdb.OpRegexNotMatch {
				if _, err := regexp.Compile(v); err != nil {
					return fmt.Errorf("tag %q: invalid regex %q: %w", ko.Key, v, err)
				}
			}
		}
	}

	return nil
}

// fluxDurationRe 校验 Flux duration 字面量（如 -1m / 5s / 30m / 24h），
// Interval 等拼接进查询的字面量必须匹配
var fluxDurationRe = regexp.MustCompile(`^-?\d+(ns|us|µs|ms|s|m|h|d|w|y)$`)

// IsValidFluxDuration 校验 s 是否为合法 Flux duration 字面量（如 -1m / 5s /
// 30m / 24h），供外部输入（如 HTTP 参数的聚合窗口）白名单校验复用；拼接进
// 查询的 duration 必须匹配，非法值既是笔误也是注入面。
func IsValidFluxDuration(s string) bool {
	return fluxDurationRe.MatchString(s)
}

// fluxEscape 转义 Flux 双引号字符串字面量中的特殊字符。Bucket/Measurement/
// tag 等外部输入直接 Sprintf 进查询时，值内含 `"` 即可逃逸出字符串并拼接
// 任意 Flux 表达式（等价 SQL 注入），此处统一转义。
func fluxEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// fluxRegexEscape 把已通过编译校验的 pattern 安全嵌入 Flux 正则字面量 /.../。
// 正则字面量以首个未转义的 `/` 收尾且不允许裸换行：pattern 里的裸 `/` 会提前
// 终止字面量、其后内容逃逸成任意 Flux 表达式（注入面），裸换行破坏单行字面量。
// 处理规则：已转义序列（\x）原样保留（用户手写的 `\/` 语义不变）；裸 `/` 改写
// 为 `\/`、裸 \n/\r 改写为对应转义（RE2 中转义标点等价原字符，语义不变）。
// 仅在 validateTags 的 regexp.Compile 通过后调用。
func fluxRegexEscape(pattern string) string {
	var b strings.Builder
	b.Grow(len(pattern))
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' && i+1 < len(pattern) {
			b.WriteByte(pattern[i])
			i++
			b.WriteByte(pattern[i])
			continue
		}
		switch pattern[i] {
		case '/':
			b.WriteString(`\/`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteByte(pattern[i])
		}
	}
	return b.String()
}

// FluxRegexLiteral 将正则 pattern 安全嵌入为 Flux 正则字面量 /pattern/（含首尾
// 斜杠），供查询拼装方（如 deck 的 schema search）复用：已转义序列原样保留、
// 裸 `/` 改写 `\/`、裸换行改写 `\n`/`\r`，防止 pattern 提前闭合字面量逃逸成
// 任意表达式（注入面）。pattern 须先经 regexp.Compile 校验（查询构建路径
// WithTag 已内置校验）；自拼查询的调用方自行保证 pattern 合法——空 pattern
// 会产出 `//`（Flux 注释开头），非法 pattern 产出不被解析的残缺字面量。
func FluxRegexLiteral(pattern string) string {
	return "/" + fluxRegexEscape(pattern) + "/"
}

// sortedTagKeys 返回按 key、op 排序的 tag 条件键，保证查询产物确定性。
func sortedTagKeys(tags map[tagOpKey][]string) []tagOpKey {
	keys := make([]tagOpKey, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Key != keys[j].Key {
			return keys[i].Key < keys[j].Key
		}
		return keys[i].Op < keys[j].Op
	})
	return keys
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

// getAggregateFn 返回拼进 aggregateWindow 的聚合函数字面量；未指定（零值）时
// 缺省 mean，保持与历史查询产物一致。合法性由 validateAggregateFn 保证。
func (t *TsdbQueryOptions) getAggregateFn() string {
	if t.AggregateFn == "" {
		return string(tsdb.AggMean)
	}
	return string(t.AggregateFn)
}

func (t *TsdbQueryOptions) getTags() string {
	if len(t.tags) == 0 {
		return ""
	}

	// sortedTagKeys 消除 map 遍历乱序，查询产物（日志/单测/排查）确定性
	var lines []string
	for _, ko := range sortedTagKeys(t.tags) {
		var items []string
		for _, v := range t.tags[ko] {
			switch ko.Op {
			case tsdb.OpEqual:
				items = append(items, fmt.Sprintf(`r["%s"] == "%s"`, fluxEscape(ko.Key), fluxEscape(v)))
			case tsdb.OpNotEqual:
				items = append(items, fmt.Sprintf(`r["%s"] != "%s"`, fluxEscape(ko.Key), fluxEscape(v)))
			case tsdb.OpRegexMatch:
				items = append(items, fmt.Sprintf(`r["%s"] =~ %s`, fluxEscape(ko.Key), FluxRegexLiteral(v)))
			case tsdb.OpRegexNotMatch:
				items = append(items, fmt.Sprintf(`r["%s"] !~ %s`, fluxEscape(ko.Key), FluxRegexLiteral(v)))
			}
		}

		// 多值语义按 De Morgan 对齐：等值/正则命中为"任一命中"用 or，
		// 不等/不命中为"排除所有"用 and
		joiner := " or "
		if ko.Op == tsdb.OpNotEqual || ko.Op == tsdb.OpRegexNotMatch {
			joiner = " and "
		}

		lines = append(lines, fmt.Sprintf("|> filter(fn: (r) => %s)", strings.Join(items, joiner)))
	}

	return strings.Join(lines, "\n")
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

// WithTag 追加一个 tag 过滤条件。同 key+op 的多次调用合并 values；不同 op
// 各自成一条 filter（pipeline 内 AND）。参数不在这里报错：空 key、空 values、
// 空串 value、未知 op、非法正则统一由 GetQuery/Query 出口的 validateTags
// 校验并返回错误，保证不生成非法查询。
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

// WithAggregateFn 指定 aggregateWindow 聚合函数（mean/max/min/last/sum/count，
// 见 tsdb.AggregateFn 常量）；未指定或缺省零值时用 mean。未知值由
// GetQuery/Query 出口校验报错。
func WithAggregateFn(fn tsdb.AggregateFn) func(*TsdbQueryOptions) {
	return func(opts *TsdbQueryOptions) {
		opts.AggregateFn = fn
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
