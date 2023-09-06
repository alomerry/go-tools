package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/tealeg/xlsx/v3"
	"testing"
)

func TestCh2Inx(t *testing.T) {
	data := []string{
		"A", "B", "a", "AA", "AAA", "AB", "AC", "I",
	}
	result := []int32{
		0, 1, 0, 26, 702, 27, 28, 8,
	}
	for i := range data {
		assert.Equalf(t, result[i], ColCh2Inx(data[i]), data[i])
		assert.Equal(t, result[i], int32(xlsx.ColLettersToIndex(data[i])), data[i])
	}
}
