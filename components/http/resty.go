package http

import (
	"context"
	"sync"

	"github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/http/opts/setup"
	resty2 "github.com/alomerry/go-tools/utils/resty"
	"github.com/go-resty/resty/v2"
)

var (
	once     = sync.Once{}
	instance = &restyClient{}
)

type Client interface {
	Post(ctx context.Context, url string, opts ...req.Opt) (Response, error)
	Get(ctx context.Context, url string, opts ...req.Opt) (Response, error)
	Close(ctx context.Context) error
}

type Response interface {
	String() string
	Bytes() []byte
	Status() string
	StatusCode() int
}

type restyClient struct {
	client *resty.Client
}

type restyResponse struct {
	*resty.Response
}

func (r *restyResponse) Bytes() []byte {
	return r.Body()
}

func NewHttpClient(opts ...setup.Opt) Client {
	rc := resty.New()
	rc.AddRetryCondition(resty2.DefaultRetryCondition)

	rc.OnBeforeRequest(resty2.DefaultRequestMiddleware)
	rc.OnAfterResponse(resty2.DefaultResponseMiddleware)

	for _, opt := range opts {
		opt(rc)
	}

	return &restyClient{
		client: rc,
	}
}

func GetClient() Client {
	once.Do(func() {
		instance = NewHttpClient().(*restyClient)
	})

	return instance
}

func (r *restyClient) Post(ctx context.Context, url string, opts ...req.Opt) (Response, error) {
	request := r.client.R()
	request.SetContext(ctx)

	for _, opt := range opts {
		opt(request)
	}

	result, err := request.Post(url)
	return &restyResponse{result}, err
}

func (r *restyClient) Get(ctx context.Context, url string, opts ...req.Opt) (Response, error) {
	request := r.client.R()
	request.SetContext(ctx)

	for _, opt := range opts {
		opt(request)
	}

	result, err := request.Get(url)
	return &restyResponse{result}, err
}

func (r *restyClient) Close(ctx context.Context) error {
	return nil
}
