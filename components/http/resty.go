package http

import (
	"context"
	stdhttp "net/http"
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
	// Header 上游响应头（deliveries 游标分页读 Link 头用；restyResponse 经内嵌
	// *resty.Response 直接满足）。
	Header() stdhttp.Header
}

type restyClient struct {
	client *resty.Client
}

type restyResponse struct {
	*resty.Response
}

// 编译期断言：restyResponse 须完整实现 Response（Header 经内嵌
// *resty.Response 满足；接口新增方法时此处先红，避免运行时才暴露）。
var _ Response = (*restyResponse)(nil)

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
	if err != nil {
		// err 非 nil 时 result 可能为 nil，包装成非 nil 接口会让调用方误判可用
		return nil, err
	}
	return &restyResponse{result}, nil
}

func (r *restyClient) Get(ctx context.Context, url string, opts ...req.Opt) (Response, error) {
	request := r.client.R()
	request.SetContext(ctx)

	for _, opt := range opts {
		opt(request)
	}

	result, err := request.Get(url)
	if err != nil {
		return nil, err
	}
	return &restyResponse{result}, nil
}

func (r *restyClient) Close(ctx context.Context) error {
	return nil
}
