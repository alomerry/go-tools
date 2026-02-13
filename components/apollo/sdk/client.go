package sdk

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Client handles communication with the Apollo Open API.
type Client struct {
	client *resty.Client

	// Services used for communicating with different parts of the Apollo Open API.
	Apps       *AppsService
	Clusters   *ClustersService
	Namespaces *NamespacesService
	Items      *ItemsService
	Releases   *ReleasesService
}

// NewClient returns a new Apollo Open API client.
// baseURL: The base URL of the Apollo Portal, e.g., "http://localhost:8070"
// token: The API token obtained from Apollo Portal
func NewClient(baseURL, token string) *Client {
	r := resty.New()
	r.SetBaseURL(baseURL)
	r.SetHeader("Authorization", token)
	r.SetHeader("Content-Type", "application/json;charset=UTF-8")

	c := &Client{
		client: r,
	}

	c.Apps = &AppsService{client: c}
	c.Clusters = &ClustersService{client: c}
	c.Namespaces = &NamespacesService{client: c}
	c.Items = &ItemsService{client: c}
	c.Releases = &ReleasesService{client: c}

	return c
}

// Do sends an API request and returns the API response.
// The API response is JSON decoded and stored in the value pointed to by v,
// or returned as an error if an API error has occurred.
func (c *Client) Do(ctx context.Context, method, path string, body interface{}, result interface{}) (*resty.Response, error) {
	req := c.client.R().SetContext(ctx)

	if body != nil {
		req.SetBody(body)
	}

	if result != nil {
		req.SetResult(result)
	}

	req.SetError(&ErrorResponse{})

	var resp *resty.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = req.Get(path)
	case http.MethodPost:
		resp, err = req.Post(path)
	case http.MethodPut:
		resp, err = req.Put(path)
	case http.MethodDelete:
		resp, err = req.Delete(path)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		if apiErr, ok := resp.Error().(*ErrorResponse); ok && apiErr.Message != "" {
			apiErr.Status = resp.StatusCode()
			return resp, apiErr
		}
		return resp, fmt.Errorf("api error: %s, body: %s", resp.Status(), string(resp.Body()))
	}

	return resp, nil
}
