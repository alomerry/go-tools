package json

import (
	"fmt"

	sonic "github.com/bytedance/sonic"
)

func Marshal(v any) string {
	str, err := sonic.MarshalString(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return str
}

func MarshalE(v any) (string, error) {
	return sonic.MarshalString(v)
}
