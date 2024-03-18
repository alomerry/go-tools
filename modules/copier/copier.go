package copier

import (
	"database/sql"
	"reflect"
)

type Converter func(from reflect.Value, toType reflect.Type) (reflect.Value, error)

type Transformer map[string]interface{}

type FieldKey string

type Mapper interface {
	From(fromValue interface{}) CopyCommand

	RegisterConverter(matcher TypeMatcher, converter Converter) Mapper

	RegisterConverterFunc(matcher TypeMatcherFunc, converter Converter) Mapper

	RegisterResetDiffField(diffFields []DiffFieldPair) Mapper

	RegisterIgnoreTargetFields(targetFieldKeys []FieldKey) Mapper

	RegisterTransformer(transformer Transformer) Mapper

	Install(Module) Mapper
}

type Module func(Mapper)

type TypeMatcher interface {
	Matches(Target) bool
}

type TypeMatcherFunc func(Target) bool

func (f TypeMatcherFunc) Matches(target Target) bool {
	return f(target)
}

type Target struct {
	From reflect.Type
	To   reflect.Type
}

func (t Target) Matches(target Target) bool {
	return t == target
}

type DiffFieldPair struct {
	Origin  string
	Targets []string
}

type CopyCommand interface {
	CopyTo(toValue interface{}) error
}

func set(to, from reflect.Value) bool {
	if from.IsValid() {
		if to.Kind() == reflect.Ptr {
			// set `to` to nil if from is nil
			if from.Kind() == reflect.Ptr && from.IsNil() {
				to.Set(reflect.Zero(to.Type()))
				return true
			} else if to.IsNil() {
				to.Set(reflect.New(to.Type().Elem()))
			}
			to = to.Elem()
		}

		if from.Type().ConvertibleTo(to.Type()) {
			to.Set(from.Convert(to.Type()))
		} else if scanner, ok := to.Addr().Interface().(sql.Scanner); ok {
			err := scanner.Scan(from.Interface())
			if err != nil {
				return false
			}
		} else if from.Kind() == reflect.Ptr {
			return set(to, from.Elem())
		} else {
			return false
		}
	}
	return true
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func indirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

func deepFields(reflectType reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	if reflectType = indirectType(reflectType); reflectType.Kind() == reflect.Struct {
		for i := 0; i < reflectType.NumField(); i++ {
			v := reflectType.Field(i)
			if v.Anonymous {
				fields = append(fields, deepFields(v.Type)...)
			} else {
				fields = append(fields, v)
			}
		}
	}

	return fields
}

func indirectAsNonNil(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return indirectAsNonNil(v.Elem())
	}

	return v
}
