package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testVars struct {
	Name  string
	Value *string
	Age   int
}

func TestVars(t *testing.T) {
	var (
		tmp = testVars{
			Name:  "init",
			Value: new(string),
			Age:   50,
		}
	)

	*tmp.Value = "init str"

	t.Run("test set json", func(t *testing.T) {
		var (
			jsonStr = `{"name": "123", "value": "123 str", "age": 5}`
		)

		assert.Equal(t, "init", tmp.Name)
		assert.Equal(t, "init str", *tmp.Value)
		assert.Equal(t, 50, tmp.Age)

		SetJson(jsonStr, reflect.ValueOf(&tmp).Elem())
		assert.Equal(t, "123", tmp.Name)
		assert.Equal(t, "123 str", *tmp.Value)
		assert.Equal(t, 5, tmp.Age)
	})
}
