package tar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUntar(t *testing.T) {
	//err := UnTar("/Users/alomerry/workspace/go/go-tools/output/2417779048_full_20251022_230101.sql.gz", "/Users/alomerry/workspace/go/go-tools/output/tmp/")
	//assert.Nil(t, err)

	err := UnTar("/Users/alomerry/workspace/go/go-tools/output/markdowns.tar.gz", "/Users/alomerry/workspace/go/go-tools/output/tmp/")
	assert.Nil(t, err)
}
