package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncCaller_Call(t *testing.T) {
	var (
		fc = NewFuncCaller("test")
	)

	fc.Register("demo1", demo1)
	fc.Register("demo2", demo2)
	fc.Register("new", newCallerTest)

	res, err := fc.Call("demo1", "test1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "demo1-test1", res[0].String())

	res, err = fc.Call("demo2", "test2")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "demo2-test2", res[0].String())

	res, err = fc.Call("new", "obj")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))

	obj := res[0].Interface()
	res, err = CallMethodByName(obj, "Do", "sleep")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "obj-sleep", res[0].String())

	/*res, err = CallMethodByName(obj, "do", "eat")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "obj-eat", res[0].String())*/
}

func demo1(name string) string {
	return fmt.Sprintf("%s-%s", "demo1", name)
}

func demo2(name string) string {
	return fmt.Sprintf("%s-%s", "demo2", name)
}

type waitCaller interface {
	do(string) string
	Do(string) string
}

type callerTest struct {
	name string
}

func (c *callerTest) do(doWhat string) string {
	return fmt.Sprintf("%s-%s", c.name, doWhat)
}

func (c *callerTest) Do(doWhat string) string {
	return fmt.Sprintf("%s-%s", c.name, doWhat)
}

func newCallerTest(name string) waitCaller {
	return &callerTest{name: name}
}
