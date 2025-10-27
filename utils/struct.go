package utils

import (
	"errors"
	"fmt"
	"reflect"
)

func CallMethodByName(receiver any, methodName string, args ...any) ([]reflect.Value, error) {
	objVal := reflect.ValueOf(receiver)
	method := objVal.MethodByName(methodName)

	if !method.IsValid() {
		ptr := reflect.New(reflect.TypeOf(receiver))
		elem := ptr.Elem()
		elem.Set(reflect.ValueOf(receiver))
		method = ptr.MethodByName(methodName)
	}

	if !method.IsValid() {
		return nil, errors.New(fmt.Sprintf("Method [%s] not found", methodName))
	}

	params := make([]reflect.Value, len(args))
	for i, arg := range args {
		argType := method.Type().In(i)
		argVal := reflect.ValueOf(arg)
		if !argVal.Type().AssignableTo(argType) {
			argVal = argVal.Convert(argType)
		}
		params[i] = argVal
	}

	return method.Call(params), nil
}
