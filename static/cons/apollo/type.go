package apollo

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type KeyType int

const (
	KeyTypeString = 0
	KeyTypeNumber = 1
	KeyTypeBool   = 2
	KeyTypeJson   = 3
)

var keyTypes = []KeyType{
	KeyTypeString, KeyTypeNumber, KeyTypeBool, KeyTypeJson,
}

func ToKeyType(s any) KeyType {
	var k int
	switch s.(type) {
	case string, int, int32, int64, uint, uint32, uint64:
		k = cast.ToInt(s)
	default:
		logrus.Panicf("invalid type %T", s)
	}

	for i := range keyTypes {
		if keyTypes[i] == KeyType(k) {
			return keyTypes[i]
		}
	}

	return -1
}
