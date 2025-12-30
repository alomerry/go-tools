package def

import (
	"testing"

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
