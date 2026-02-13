package req

import "github.com/go-resty/resty/v2"

type Opt func(*resty.Request)

func WithPathParam(param map[string]string) Opt {
	return func(request *resty.Request) {
		request.SetPathParams(param)
	}
}

func WithQueryParam(param map[string]string) Opt {
	return func(request *resty.Request) {
		request.SetQueryParams(param)
	}
}

func WithAuthToken(authToken string) Opt {
	return func(request *resty.Request) {
		request.SetAuthToken(authToken)
	}
}

func WithAuthSchema(scheme string) Opt {
	return func(request *resty.Request) {
		request.SetAuthScheme(scheme)
	}
}

func WithBody(body interface{}) Opt {
	return func(request *resty.Request) {
		request.SetBody(body)
	}
}
