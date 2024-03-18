package copier

import (
	"reflect"
	"time"
)

const RFC3339Mili = "2006-01-02T15:04:05.999Z07:00"

func RFC3339Convertor(m Mapper) {
	registerTimeToRFC3339Converter(m)
}

func registerTimeToRFC3339Converter(m Mapper) {
	m.RegisterConverter(
		Target{
			From: reflect.TypeOf(time.Time{}),
			To:   reflect.TypeOf(""),
		},
		func(from reflect.Value, _ reflect.Type) (reflect.Value, error) {
			if timeValue, ok := from.Interface().(time.Time); ok {
				if timeValue.Unix() > 0 {
					return reflect.ValueOf(timeValue.Format(RFC3339Mili)), nil
				} else {
					return reflect.ValueOf(""), nil
				}
			}
			return from, nil
		},
	)
}
