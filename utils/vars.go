package utils

import (
	"encoding/json"
	"reflect"
)

func SetJson(val string, dist reflect.Value) error {
	midVal := reflect.New(dist.Type())
	err := json.Unmarshal([]byte(val), midVal.Interface())
	if err != nil {
		return err
	}

	dist.Set(midVal.Elem())
	return nil
}
