package proto

import (
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/spf13/cast"
)

func AnyToStructValue(v any) (*_struct.Value, error) {
	if v == nil {
		return &_struct.Value{
			Kind: &_struct.Value_NullValue{NullValue: _struct.NullValue_NULL_VALUE},
		}, nil
	}

	switch val := v.(type) {
	case bool:
		return &_struct.Value{
			Kind: &_struct.Value_BoolValue{BoolValue: val},
		}, nil
	case string:
		return &_struct.Value{
			Kind: &_struct.Value_StringValue{StringValue: val},
		}, nil
	case int, int8, int16, int32, int64:
		return &_struct.Value{
			Kind: &_struct.Value_NumberValue{NumberValue: cast.ToFloat64(val)},
		}, nil
	case uint, uint8, uint16, uint32, uint64:
		return &_struct.Value{
			Kind: &_struct.Value_NumberValue{NumberValue: cast.ToFloat64(val)},
		}, nil
	case float32, float64:
		return &_struct.Value{
			Kind: &_struct.Value_NumberValue{NumberValue: cast.ToFloat64(val)},
		}, nil
	default:
		return &_struct.Value{
			Kind: &_struct.Value_StringValue{StringValue: cast.ToString(val)},
		}, nil
	}
}
