package setup

import (
	"time"

	"resty.dev/v3"
)

type Opt func(*resty.Client)

func WithAuthSchema(scheme string) Opt {
	return func(rc *resty.Client) {
		rc.SetAuthScheme(scheme)
	}
}

func WithAuthToken(authToken string) Opt {
	return func(rc *resty.Client) {
		rc.SetAuthToken(authToken)
	}
}

func WithRequestMiddleware(middleware resty.RequestMiddleware) Opt {
	return func(rc *resty.Client) {
		rc.AddRequestMiddleware(middleware)
	}
}

func WithContentType(contentType string) Opt {
	return func(rc *resty.Client) {
		rc.SetHeader("Content-Type", contentType)
	}
}

func WithAccept(accept string) Opt {
	return func(rc *resty.Client) {
		rc.SetHeader("Accept", accept)
	}
}

func WithResponseMiddleware(middleware resty.ResponseMiddleware) Opt {
	return func(rc *resty.Client) {
		rc.AddResponseMiddleware(middleware)
	}
}

func WithHeader(k, v string) Opt {
	return func(rc *resty.Client) {
		rc.SetHeader(k, v)
	}
}

func WithTimeout(timeout time.Duration) Opt {
	return func(rc *resty.Client) {
		rc.SetTimeout(timeout)
	}
}

func WithRetryCount(count int) Opt {
	return func(rc *resty.Client) {
		rc.SetRetryCount(count)
	}
}

func WithBaseURL(baseURL string) Opt {
	return func(rc *resty.Client) {
		rc.SetBaseURL(baseURL)
	}
}

func Debug(debug bool) Opt {
	return func(rc *resty.Client) {
		rc.SetDebug(debug)
	}
}
