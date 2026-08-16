package internal

import (
	"io"
	"strings"
	"testing"

	"github.com/alomerry/go-tools/components/tsdb/def"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseQueryResultsFieldName is a response-level regression test for the
// default_count root cause: the series assembler must derive the series name and
// Columns from the real Flux `_field` value ("usage") present in the response
// CSV, never from the "default_count" fallback.
//
// The CSV mirrors what InfluxDB returns for the fixed query
// `group(columns: ["_field"]) |> aggregateWindow(every: 5s, fn: mean, createEmpty: false)`,
// i.e. every table carries a `_field` group column so record.Field() is non-empty.
func TestParseQueryResultsFieldName(t *testing.T) {
	t.Run("no tag group keeps real field name", func(t *testing.T) {
		csv := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string
#group,false,false,true,true,false,true,true,true
#default,_result,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement
,,0,2026-08-16T00:00:00Z,2026-08-16T00:05:00Z,2026-08-16T00:00:00Z,0.5,usage,cpu.usage
,,0,2026-08-16T00:00:00Z,2026-08-16T00:05:00Z,2026-08-16T00:01:00Z,0.6,usage,cpu.usage
`
		options := new(def.TsdbQueryOptions)
		options.Apply(
			def.WithBucket("homelab"),
			def.WithMeasurement("cpu.usage"),
			def.WithFields("usage"),
		)

		res, err := parseQueryResults(options, newQueryTableResult(t, csv))
		require.NoError(t, err)
		require.Len(t, res, 1)

		got := res[0]
		assert.Equal(t, "usage", got.Name)
		assert.NotContains(t, got.Name, "default_count")
		assert.Equal(t, []string{"time", "usage"}, got.Columns)
		require.Len(t, got.Values, 2)
		assert.Equal(t, []any{int64(1786838400), 0.5}, got.Values[0])
	})

	t.Run("tag group keeps real field name and tag", func(t *testing.T) {
		csv := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#group,false,false,true,true,false,true,true,true,true
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,service
,,0,2026-08-16T00:00:00Z,2026-08-16T00:05:00Z,2026-08-16T00:00:00Z,0.5,usage,cpu.usage,svc-a
`
		options := new(def.TsdbQueryOptions)
		options.Apply(
			def.WithBucket("homelab"),
			def.WithMeasurement("cpu.usage"),
			def.WithFields("usage"),
			def.WithGroup("service"),
		)

		res, err := parseQueryResults(options, newQueryTableResult(t, csv))
		require.NoError(t, err)
		require.Len(t, res, 1)

		got := res[0]
		assert.Equal(t, "svc-a-usage", got.Name)
		assert.NotContains(t, got.Name, "default_count")
		assert.Equal(t, map[string]string{"service": "svc-a"}, got.Tags)
		assert.Equal(t, []string{"time", "usage"}, got.Columns)
	})

	t.Run("multiple fields produce one series per field", func(t *testing.T) {
		csv := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string
#group,false,false,true,true,false,true,true,true
#default,_result,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement
,,0,2026-08-16T00:00:00Z,2026-08-16T00:05:00Z,2026-08-16T00:00:00Z,0.5,usage,cpu.usage

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string
#group,false,false,true,true,false,true,true,true
#default,_result,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement
,,1,2026-08-16T00:00:00Z,2026-08-16T00:05:00Z,2026-08-16T00:00:00Z,3,count,cpu.usage
`
		options := new(def.TsdbQueryOptions)
		options.Apply(
			def.WithBucket("homelab"),
			def.WithMeasurement("cpu.usage"),
			def.WithFields("usage", "count"),
		)

		res, err := parseQueryResults(options, newQueryTableResult(t, csv))
		require.NoError(t, err)
		require.Len(t, res, 2)

		names := make([]string, 0, len(res))
		for _, s := range res {
			names = append(names, s.Name)
		}
		assert.ElementsMatch(t, []string{"usage", "count"}, names)
		for _, s := range res {
			assert.NotContains(t, s.Name, "default_count")
			assert.Equal(t, []string{"time", s.Name}, s.Columns)
		}
	})
}

// TestParseQueryResultsDefaultCountFallback pins the failure mode the
// group-by-_field fix addresses: when the response CSV carries no `_field`
// column (the pre-fix `group(columns: [])` output), record.Field() is empty and
// the assembler falls back to "default_count". The fixed query path must never
// reach this fallback; keeping the test documents the root cause.
func TestParseQueryResultsDefaultCountFallback(t *testing.T) {
	csv := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string
#group,false,false,true,true,false,true,true
#default,_result,,,,,,
,result,table,_start,_stop,_time,_value,_measurement
,,0,2026-08-16T00:00:00Z,2026-08-16T00:05:00Z,2026-08-16T00:00:00Z,0.5,cpu.usage
`
	options := new(def.TsdbQueryOptions)
	options.Apply(
		def.WithBucket("homelab"),
		def.WithMeasurement("cpu.usage"),
		def.WithFields("usage"),
	)

	res, err := parseQueryResults(options, newQueryTableResult(t, csv))
	require.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, "default_count", res[0].Name)
	assert.Equal(t, []string{"time", "default_count"}, res[0].Columns)
}

func newQueryTableResult(t *testing.T, csv string) *api.QueryTableResult {
	t.Helper()
	return api.NewQueryTableResult(io.NopCloser(strings.NewReader(csv)))
}
