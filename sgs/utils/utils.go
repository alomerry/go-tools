package utils

import (
	"fmt"
	"github.com/spf13/cast"
	"math"
	"strings"
	"unicode"
)

func ColCh2Inx(col string) int32 {
	col = strings.ToLower(col)
	res, err := cast.ToInt32E(col)
	if err == nil {
		return res
	}
	res = 0
	for i := len(col) - 1; i >= 0; i-- {
		if !unicode.IsLetter(rune(col[i])) {
			panic(fmt.Sprintf("[%v] is not a letter", col[i]))
		}
		res += int32(col[i]-'a'+1) * int32(math.Pow(float64(26), float64(len(col)-1-i)))
	}
	return res - 1
}

func GetUniqueColRange(from, to int) []int {
	var result = []int{from}
	if from == to {
		return result
	}
	i := result[len(result)-1] + 1
	for i <= to {
		result = append(result, i)
		i++
	}
	return result
}
