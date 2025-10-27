package utils

import (
	"fmt"
	"reflect"
)

type FuncCaller struct {
	name     string
	registry map[string]reflect.Value
}

func NewFuncCaller(name string) *FuncCaller {
	return &FuncCaller{
		name:     name,
		registry: make(map[string]reflect.Value),
	}
}

func (f *FuncCaller) Name() string {
	return f.name
}

func (f *FuncCaller) Call(name string, args ...any) ([]reflect.Value, error) {
	fn, ok := f.registry[name]
	if !ok {
		return nil, fmt.Errorf("function %s not registered", name)
	}

	fnType := fn.Type()
	if fnType.NumIn() != len(args) {
		return nil, fmt.Errorf("function %s expects %d arguments, got %d",
			name, fnType.NumIn(), len(args))
	}

	params := make([]reflect.Value, len(args))
	for i, arg := range args {
		argType := fnType.In(i)
		argVal := reflect.ValueOf(arg)
		if !argVal.Type().AssignableTo(argType) {
			argVal = argVal.Convert(argType)
		}
		params[i] = argVal
	}

	return fn.Call(params), nil
}

func (f *FuncCaller) Register(name string, fn any) {
	f.registry[name] = reflect.ValueOf(fn)
}
