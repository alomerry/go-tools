package def

import (
	"testing"
	"time"

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
