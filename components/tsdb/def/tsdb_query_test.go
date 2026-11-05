package def

import (
	"testing"
	"time"

	"github.com/alomerry/go-tools/static/cons/tsdb"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	query := new(TsdbQueryOptions)

	options := append([]func(*TsdbQueryOptions){},
		WithBucket("homelab"),
		WithMeasurement("cpu.usage"),
		WithFields("usage"),
	)

	query.Apply(options...)

	str, err := query.GetQuery()
	assert.NoError(t, err)
	t.Logf("%v", str)
}

func TestGetQueryRangeTimeLiteral(t *testing.T) {
	start := time.Date(2026, 8, 16, 11, 34, 27, 0, time.FixedZone("CST", 8*3600))
	end := time.Date(2026, 8, 16, 17, 34, 27, 0, time.FixedZone("CST", 8*3600))

	// +08:00 wall-clock time parsed in a fixed CST zone: 2026-08-15 11:34:27 CST
	// is the same instant as 2026-08-15 03:34:27 UTC. time.FixedZone keeps the
	// test self-contained (no system tzdata dependency); Asia/Shanghai has no
	// DST, so +08:00 is constant and semantically equivalent.
	parsedStart, err := time.ParseInLocation("2006-01-02 15:04:05", "2026-08-15 11:34:27", time.FixedZone("CST", 8*3600))
	assert.NoError(t, err)

	tests := []struct {
		name      string
		start     *time.Time
		end       *time.Time
		wantRange string
	}{
		{
			name:      "explicit start and end convert to UTC and emit unquoted Z literals",
			start:     &start,
			end:       &end,
			wantRange: `|> range(start: 2026-08-16T03:34:27Z, stop: 2026-08-16T09:34:27Z)`,
		},
		{
			name:      "start parsed from fixed CST wall clock emits UTC Z literal",
			start:     &parsedStart,
			wantRange: `|> range(start: 2026-08-15T03:34:27Z, stop: now())`,
		},
		{
			name:      "only end set converts to UTC Z literal and keeps start keyword",
			end:       &end,
			wantRange: `|> range(start: -1m, stop: 2026-08-16T09:34:27Z)`,
		},
		{
			name:      "nil start and end use unquoted flux keywords",
			wantRange: `|> range(start: -1m, stop: now())`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []func(*TsdbQueryOptions){
				WithBucket("homelab"),
				WithMeasurement("cpu.usage"),
				WithFields("usage"),
			}
			if tt.start != nil {
				opts = append(opts, WithStart(*tt.start))
			}
			if tt.end != nil {
				opts = append(opts, WithEnd(*tt.end))
			}

			query := new(TsdbQueryOptions)
			query.Apply(opts...)

			str, err := query.GetQuery()
			assert.NoError(t, err)
			assert.Contains(t, str, tt.wantRange)
		})
	}
}

func TestGetQueryGroupEmptyKeepsField(t *testing.T) {
	query := new(TsdbQueryOptions)
	query.Apply(
		WithBucket("homelab"),
		WithMeasurement("cpu.usage"),
		WithFields("usage"),
	)

	str, err := query.GetQuery()
	assert.NoError(t, err)
	assert.Contains(t, str, `|> group(columns: ["_field"])`)
	assert.NotContains(t, str, `|> group(columns: [])`)
}

func TestGetQueryGroupWithTagsKeepsField(t *testing.T) {
	query := new(TsdbQueryOptions)
	query.Apply(
		WithBucket("homelab"),
		WithMeasurement("cpu.usage"),
		WithFields("usage"),
		WithGroup("service"),
	)

	str, err := query.GetQuery()
	assert.NoError(t, err)
	assert.Contains(t, str, `|> group(columns: ["service","_field"])`)
}

func TestGetQueryAggregateWindowCreateEmptyFalse(t *testing.T) {
	query := new(TsdbQueryOptions)
	query.Apply(
		WithBucket("homelab"),
		WithMeasurement("cpu.usage"),
		WithFields("usage"),
	)

	str, err := query.GetQuery()
	assert.NoError(t, err)
	assert.Contains(t, str, `|> aggregateWindow(every: 5s, fn: mean, createEmpty: false)`)
}

func TestGetQueryTagFilterOps(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		op         tsdb.TagOp
		values     []string
		wantFilter string
	}{
		{
			name:       "equal single value",
			key:        "service",
			op:         tsdb.OpEqual,
			values:     []string{"deck"},
			wantFilter: `|> filter(fn: (r) => r["service"] == "deck")`,
		},
		{
			name:       "equal multi values joined with or",
			key:        "service",
			op:         tsdb.OpEqual,
			values:     []string{"deck", "gateway"},
			wantFilter: `|> filter(fn: (r) => r["service"] == "deck" or r["service"] == "gateway")`,
		},
		{
			name:       "not equal multi values joined with and",
			key:        "service",
			op:         tsdb.OpNotEqual,
			values:     []string{"deck", "gateway"},
			wantFilter: `|> filter(fn: (r) => r["service"] != "deck" and r["service"] != "gateway")`,
		},
		{
			name:       "regex match single pattern",
			key:        "host",
			op:         tsdb.OpRegexMatch,
			values:     []string{`^vm1\d+$`},
			wantFilter: `|> filter(fn: (r) => r["host"] =~ /^vm1\d+$/)`,
		},
		{
			name:       "regex match multi patterns joined with or",
			key:        "host",
			op:         tsdb.OpRegexMatch,
			values:     []string{`^a`, `^b`},
			wantFilter: `|> filter(fn: (r) => r["host"] =~ /^a/ or r["host"] =~ /^b/)`,
		},
		{
			name:       "regex not match multi patterns joined with and",
			key:        "host",
			op:         tsdb.OpRegexNotMatch,
			values:     []string{`test$`, `tmp$`},
			wantFilter: `|> filter(fn: (r) => r["host"] !~ /test$/ and r["host"] !~ /tmp$/)`,
		},
		{
			name:       "value with double quote is escaped",
			key:        "name",
			op:         tsdb.OpEqual,
			values:     []string{`a"b`},
			wantFilter: `|> filter(fn: (r) => r["name"] == "a\"b")`,
		},
		{
			name:       "tag key with double quote is escaped",
			key:        `na"me`,
			op:         tsdb.OpEqual,
			values:     []string{"v"},
			wantFilter: `|> filter(fn: (r) => r["na\"me"] == "v")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := new(TsdbQueryOptions)
			query.Apply(
				WithBucket("homelab"),
				WithMeasurement("cpu.usage"),
				WithTag(tt.key, tt.op, tt.values...),
			)

			str, err := query.GetQuery()
			assert.NoError(t, err)
			assert.Contains(t, str, tt.wantFilter)
		})
	}
}

func TestGetQueryTagFilterRegexEscaping(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		wantFilter string
	}{
		{
			name:       "slash in pattern is escaped to keep literal closed",
			pattern:    `^prod/.*`,
			wantFilter: `|> filter(fn: (r) => r["env"] =~ /^prod\/.*/)`,
		},
		{
			name:       "user escaped slash is preserved",
			pattern:    `^a\/b$`,
			wantFilter: `|> filter(fn: (r) => r["env"] =~ /^a\/b$/)`,
		},
		{
			name:       "injection via early literal close is neutralized",
			pattern:    `x/ or r["a"] == "b"`,
			wantFilter: `|> filter(fn: (r) => r["env"] =~ /x\/ or r["a"] == "b"/)`,
		},
		{
			name:       "raw newline is rewritten as regex escape",
			pattern:    "a\nb",
			wantFilter: `|> filter(fn: (r) => r["env"] =~ /a\nb/)`,
		},
		{
			name:       "escaped newline sequence is preserved",
			pattern:    `a\nb`,
			wantFilter: `|> filter(fn: (r) => r["env"] =~ /a\nb/)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := new(TsdbQueryOptions)
			query.Apply(
				WithBucket("homelab"),
				WithMeasurement("cpu.usage"),
				WithTag("env", tsdb.OpRegexMatch, tt.pattern),
			)

			str, err := query.GetQuery()
			assert.NoError(t, err)
			assert.Contains(t, str, tt.wantFilter)
		})
	}
}

func TestFluxRegexEscape(t *testing.T) {
	assert.Equal(t, `a\/b`, fluxRegexEscape(`a/b`))
	assert.Equal(t, `a\/b`, fluxRegexEscape(`a\/b`))
	assert.Equal(t, `\\`, fluxRegexEscape(`\\`))
	assert.Equal(t, `\n`, fluxRegexEscape("\n"))
	assert.Equal(t, `\r`, fluxRegexEscape("\r"))
	assert.Equal(t, "a\\", fluxRegexEscape(`a\`))
}

func TestGetQueryTagValidation(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		op      tsdb.TagOp
		values  []string
		wantErr string
	}{
		{
			name:    "empty values list rejected instead of emitting invalid flux",
			key:     "service",
			op:      tsdb.OpEqual,
			values:  []string{},
			wantErr: "at least one value is required",
		},
		{
			name:    "all empty string values rejected",
			key:     "service",
			op:      tsdb.OpEqual,
			values:  []string{"", ""},
			wantErr: "value must not be empty",
		},
		{
			name:    "any empty string value rejected",
			key:     "service",
			op:      tsdb.OpEqual,
			values:  []string{"a", ""},
			wantErr: "value must not be empty",
		},
		{
			name:    "empty key rejected",
			key:     "",
			op:      tsdb.OpEqual,
			values:  []string{"a"},
			wantErr: "tag key must not be empty",
		},
		{
			name:    "unknown op rejected",
			key:     "service",
			op:      tsdb.TagOp("~"),
			values:  []string{"a"},
			wantErr: "unsupported tag op",
		},
		{
			name:    "invalid regex rejected for match",
			key:     "env",
			op:      tsdb.OpRegexMatch,
			values:  []string{"a(b"},
			wantErr: "invalid regex",
		},
		{
			name:    "invalid regex rejected for not match",
			key:     "env",
			op:      tsdb.OpRegexNotMatch,
			values:  []string{"*x"},
			wantErr: "invalid regex",
		},
		{
			name:    "empty regex pattern rejected",
			key:     "env",
			op:      tsdb.OpRegexMatch,
			values:  []string{""},
			wantErr: "value must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := new(TsdbQueryOptions)
			query.Apply(
				WithBucket("homelab"),
				WithMeasurement("cpu.usage"),
				WithTag(tt.key, tt.op, tt.values...),
			)

			_, err := query.GetQuery()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestGetQueryTagFiltersSortedDeterministic(t *testing.T) {
	query := new(TsdbQueryOptions)
	query.Apply(
		WithBucket("homelab"),
		WithMeasurement("cpu.usage"),
		WithTag("zeta", tsdb.OpEqual, "1"),
		WithTag("alpha", tsdb.OpEqual, "2"),
		WithTag("alpha", tsdb.OpNotEqual, "3"),
	)

	str, err := query.GetQuery()
	assert.NoError(t, err)
	// 按 key、op 字典序（"!=" < "="）：同 key 的条件顺序稳定可复现
	assert.Contains(t, str, `|> filter(fn: (r) => r["alpha"] != "3")
|> filter(fn: (r) => r["alpha"] == "2")
|> filter(fn: (r) => r["zeta"] == "1")`)
}

func TestGetQueryAggregateFn(t *testing.T) {
	tests := []struct {
		name    string
		fn      tsdb.AggregateFn
		want    string
		wantErr string
	}{
		{
			name: "zero value falls back to mean",
			fn:   "",
			want: `|> aggregateWindow(every: 5s, fn: mean, createEmpty: false)`,
		},
		{
			name: "max",
			fn:   tsdb.AggMax,
			want: `|> aggregateWindow(every: 5s, fn: max, createEmpty: false)`,
		},
		{
			name: "min",
			fn:   tsdb.AggMin,
			want: `|> aggregateWindow(every: 5s, fn: min, createEmpty: false)`,
		},
		{
			name: "last",
			fn:   tsdb.AggLast,
			want: `|> aggregateWindow(every: 5s, fn: last, createEmpty: false)`,
		},
		{
			name: "sum",
			fn:   tsdb.AggSum,
			want: `|> aggregateWindow(every: 5s, fn: sum, createEmpty: false)`,
		},
		{
			name: "count",
			fn:   tsdb.AggCount,
			want: `|> aggregateWindow(every: 5s, fn: count, createEmpty: false)`,
		},
		{
			name:    "unknown fn rejected",
			fn:      tsdb.AggregateFn(`mean) or (x`),
			wantErr: "unsupported aggregate fn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := new(TsdbQueryOptions)
			query.Apply(
				WithBucket("homelab"),
				WithMeasurement("cpu.usage"),
				WithFields("usage"),
				WithAggregateFn(tt.fn),
			)

			str, err := query.GetQuery()
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Contains(t, str, tt.want)
		})
	}
}

func TestGetQueryWindowOverride(t *testing.T) {
	t.Run("manual window overrides auto interval", func(t *testing.T) {
		query := new(TsdbQueryOptions)
		query.Apply(
			WithBucket("homelab"),
			WithMeasurement("cpu.usage"),
			WithInterval("30s"),
		)

		str, err := query.GetQuery()
		assert.NoError(t, err)
		assert.Contains(t, str, `|> aggregateWindow(every: 30s, fn: mean, createEmpty: false)`)
	})

	t.Run("manual window combines with aggregate fn", func(t *testing.T) {
		query := new(TsdbQueryOptions)
		query.Apply(
			WithBucket("homelab"),
			WithMeasurement("cpu.usage"),
			WithInterval("1m"),
			WithAggregateFn(tsdb.AggMax),
		)

		str, err := query.GetQuery()
		assert.NoError(t, err)
		assert.Contains(t, str, `|> aggregateWindow(every: 1m, fn: max, createEmpty: false)`)
	})

	t.Run("invalid window literal rejected", func(t *testing.T) {
		query := new(TsdbQueryOptions)
		query.Apply(
			WithBucket("homelab"),
			WithMeasurement("cpu.usage"),
			WithInterval(`5s) or (x`),
		)

		_, err := query.GetQuery()
		assert.ErrorContains(t, err, "invalid flux duration")
	})
}

func TestFluxRegexLiteral(t *testing.T) {
	assert.Equal(t, `/^prod\/.*/`, FluxRegexLiteral(`^prod/.*`))
	assert.Equal(t, `//`, FluxRegexLiteral("")) // 空 pattern 产出 //（Flux 注释开头），查询路径由 validateTags 拒绝
	assert.Equal(t, `/a\nb/`, FluxRegexLiteral("a\nb"))
}

func TestIsValidFluxDuration(t *testing.T) {
	for _, ok := range []string{"-1m", "5s", "30m", "24h", "1w", "1y", "500ms", "100us", "100µs", "-30s"} {
		assert.True(t, IsValidFluxDuration(ok), "want valid: %q", ok)
	}
	for _, bad := range []string{"", "5", "5 s", `5s) or (x`, "abc", "1.5h"} {
		assert.False(t, IsValidFluxDuration(bad), "want invalid: %q", bad)
	}
}
