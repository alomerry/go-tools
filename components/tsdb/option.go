package tsdb

import "github.com/sirupsen/logrus"

func WithEndpoint(endpoint string) func(any) {
	return func(v any) {
		switch v.(type) {
		case *metric:
		case *meta:
			v.(*meta).endpoint = endpoint
		default:
			logrus.Errorf("not support option type: %T", v)
		}
	}
}

func WithOrg(org string) func(any) {
	return func(v any) {
		switch v.(type) {
		case *metric:
		case *meta:
			v.(*meta).org = org
		default:
			logrus.Errorf("not support option type: %T", v)
		}
	}
}

func WithToken(token string) func(any) {
	return func(v any) {
		switch v.(type) {
		case *metric:
		case *meta:
			v.(*meta).token = token
		default:
			logrus.Errorf("not support option type: %T", v)
		}
	}
}

func WithBucket(bucket string) func(any) {
	return func(v any) {
		switch v.(type) {
		case *metric:
			v.(*metric).Bucket = bucket
		case *meta:
			v.(*meta).bucket = bucket
		default:
			logrus.Errorf("not support option type: %T", v)
		}
	}
}

func WithMetric(measurement string) func(any) {
	return func(v any) {
		switch v.(type) {
		case *metric:
			v.(*metric).Measurement = measurement
		case *meta:
		default:
			logrus.Errorf("not support option type: %T", v)
		}
	}
}

func WithTag(k, v string) func(any) {
	return withTagOrField(k, v)
}

func WithTags(tags map[string]string) func(any) {
	return func(m any) {
		for k, v := range tags {
			WithTag(k, v)(m)
		}
	}
}

func WithField(k string, v any) func(any) {
	return withTagOrField(k, v)
}

func WithFields(fields map[string]any) func(any) {
	return func(m any) {
		for k, v := range fields {
			WithField(k, v)(m)
		}
	}
}

// WithFieldAny 显式将键值写入 Fields，绕过 string→Tags 的隐式映射，
// 用于需要字符串类型 field 的场景（如 event.data / problem.message）。
func WithFieldAny(k string, v any) func(any) {
	return func(m any) {
		switch m := m.(type) {
		case *metric:
			if len(k) == 0 || v == nil {
				return
			}
			m.Fields[k] = v
		case *meta:
		default:
			logrus.Errorf("not support option type: %T", m)
		}
	}
}

func withTagOrField(k string, v any) func(any) {
	return func(m any) {
		var m1 *metric
		switch m.(type) {
		case *metric:
			m1 = m.(*metric)
		case *meta:
			return
		default:
			// 目标容器类型不支持，打 m 的类型（原打 v 的类型会误导排查）
			logrus.Errorf("not support option type: %T", m)
			return
		}

		if len(k) == 0 || v == nil {
			return
		}
		switch v.(type) {
		case string:
			m1.Tags[k] = v.(string)
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64, bool:
			m1.Fields[k] = v
		}
	}
}
