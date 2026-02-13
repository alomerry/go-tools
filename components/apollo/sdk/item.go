package sdk

import (
	"context"
	"fmt"
	"net/http"
)

type ItemsService struct {
	client *Client
}

type Item struct {
	Key                        string `json:"key"`
	Value                      string `json:"value"`
	Comment                    string `json:"comment"`
	DataChangeCreatedBy        string `json:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string `json:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string `json:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string `json:"dataChangeLastModifiedTime"`
}

type PageDTO struct {
	Content []*Item `json:"content"`
	Page    int     `json:"page"`
	Size    int     `json:"size"`
	Total   int     `json:"total"`
}

func (s *ItemsService) Get(ctx context.Context, env, appID, clusterName, namespaceName, key string) (*Item, error) {
	path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s", env, appID, clusterName, namespaceName, key)
	var result Item
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (s *ItemsService) Create(ctx context.Context, env, appID, clusterName, namespaceName string, item *Item) (*Item, error) {
	path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items", env, appID, clusterName, namespaceName)
	var result Item
	_, err := s.client.Do(ctx, http.MethodPost, path, item, &result)
	return &result, err
}

func (s *ItemsService) Update(ctx context.Context, env, appID, clusterName, namespaceName string, item *Item) (*Item, error) {
	path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s", env, appID, clusterName, namespaceName, item.Key)
	var result Item
	_, err := s.client.Do(ctx, http.MethodPut, path, item, &result)
	return &result, err
}

func (s *ItemsService) Delete(ctx context.Context, env, appID, clusterName, namespaceName, key, operator string) error {
	path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s?operator=%s", env, appID, clusterName, namespaceName, key, operator)
	_, err := s.client.Do(ctx, http.MethodDelete, path, nil, nil)
	return err
}

func (s *ItemsService) List(ctx context.Context, env, appID, clusterName, namespaceName string, page, size int) (*PageDTO, error) {
	path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items?page=%d&size=%d", env, appID, clusterName, namespaceName, page, size)
	var result PageDTO
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}
