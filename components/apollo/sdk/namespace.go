package sdk

import (
	"context"
	"fmt"
	"net/http"
)

type NamespacesService struct {
	client *Client
}

type Namespace struct {
	AppID                      string `json:"appId"`
	ClusterName                string `json:"clusterName"`
	NamespaceName              string `json:"namespaceName"`
	Comment                    string `json:"comment"`
	Format                     string `json:"format"`
	IsPublic                   bool   `json:"isPublic"`
	Items                      []Item `json:"items"`
	DataChangeCreatedBy        string `json:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string `json:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string `json:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string `json:"dataChangeLastModifiedTime"`
}

type NamespaceLock struct {
	NamespaceName string `json:"namespaceName"`
	IsLocked      bool   `json:"isLocked"`
	LockedBy      string `json:"lockedBy"`
}

func (s *NamespacesService) GetAll(ctx context.Context, info NamespaceQuery) ([]*Namespace, error) {
  path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces", info.GetEnv(), info.GetAppId(), info.GetCluster())
	var result []*Namespace
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return result, err
}

func (s *NamespacesService) Get(ctx context.Context, info FullQuery) (*Namespace, error) {
  path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s", info.GetEnv(), info.GetAppId(), info.GetCluster(), info.GetNamespace())
	var result Namespace
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (s *NamespacesService) Create(ctx context.Context, info NamespaceQuery, namespace *Namespace) (*Namespace, error) {
  path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces", info.GetEnv(), info.GetAppId(), info.GetCluster())
	var result Namespace
	_, err := s.client.Do(ctx, http.MethodPost, path, namespace, &result)
	return &result, err
}

func (s *NamespacesService) GetLock(ctx context.Context, info FullQuery) (*NamespaceLock, error) {
  path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/lock", info.GetEnv(), info.GetAppId(), info.GetCluster(), info.GetNamespace())
	var result NamespaceLock
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}
