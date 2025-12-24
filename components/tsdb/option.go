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

func withTagOrField(k string, v any) func(any) {
	return func(m any) {
		var m1 *metric
		switch m.(type) {
		case *metric:
			m1 = m.(*metric)
		case *meta:
			return
		default:
			logrus.Errorf("not support option type: %T", v)
			return
		}

		if len(k) == 0 || v == nil {
			return
		}
		switch v.(type) {
		case string:
			m1.Tags[k] = v.(string)
		case int64, float64, uint64, float32, uint32, int32:
			m1.Fields[k] = v
		}
	}
}
