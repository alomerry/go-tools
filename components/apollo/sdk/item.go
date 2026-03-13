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
  Key   string `json:"key"`
  Type  int    `json:"type"`
  Value string `json:"value"`
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

func (s *ItemsService) Get(ctx context.Context, info FullQuery, key string) (*Item, error) {
  path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s", info.GetEnv(), info.GetAppId(), info.GetCluster(), info.GetNamespace(), key)
	var result Item
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (s *ItemsService) Create(ctx context.Context, info FullQuery, item *Item) (*Item, error) {
  path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items", info.GetEnv(), info.GetAppId(), info.GetCluster(), info.GetNamespace())
	var result Item
	_, err := s.client.Do(ctx, http.MethodPost, path, item, &result)
	return &result, err
}

func (s *ItemsService) Update(ctx context.Context, info FullQuery, item *Item) (*Item, error) {
  path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s", info.GetEnv(), info.GetAppId(), info.GetCluster(), info.GetNamespace(), item.Key)
	var result Item
	_, err := s.client.Do(ctx, http.MethodPut, path, item, &result)
	return &result, err
}

func (s *ItemsService) Delete(ctx context.Context, info FullQuery, key, operator string) error {
  path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s?operator=%s", info.GetEnv(), info.GetAppId(), info.GetCluster(), info.GetNamespace(), key, operator)
	_, err := s.client.Do(ctx, http.MethodDelete, path, nil, nil)
	return err
}

func (s *ItemsService) List(ctx context.Context, info FullQuery, page, size int) (*PageDTO, error) {
  path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items?page=%d&size=%d", info.GetEnv(), info.GetAppId(), info.GetCluster(), info.GetNamespace(), page, size)
	var result PageDTO
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}
