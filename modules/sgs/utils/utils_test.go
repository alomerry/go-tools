package utils

import (
	rand2 "math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tealeg/xlsx/v3"
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

func TestColCh2Inx1(t *testing.T) {
	map1 := make(map[string]int32)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000000000; i++ {
			time.Sleep(time.Microsecond * time.Duration(rand2.Int31n(50)))
			for k, v := range map1 {
				_, _ = k, v
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000000000; i++ {
			time.Sleep(time.Microsecond * time.Duration(rand2.Int31n(50)))
			if time.Now().Unix()%int64(rand2.Int31n(100)+1) == 0 {
				map1["a"] = 1
			}
		}
	}()

	wg.Wait()
}
