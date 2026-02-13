package sdk

import (
	"context"
	"fmt"
	"net/http"
)

type ClustersService struct {
	client *Client
}

type Cluster struct {
	Name                       string `json:"name"`
	AppID                      string `json:"appId"`
	ParentClusterID            string `json:"parentClusterId"`
	DataChangeCreatedBy        string `json:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string `json:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string `json:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string `json:"dataChangeLastModifiedTime"`
}

func (s *ClustersService) Get(ctx context.Context, env, appID, clusterName string) (*Cluster, error) {
	path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s", env, appID, clusterName)
	var result Cluster
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (s *ClustersService) Create(ctx context.Context, env, appID string, cluster *Cluster) (*Cluster, error) {
	path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters", env, appID)
	var result Cluster
	_, err := s.client.Do(ctx, http.MethodPost, path, cluster, &result)
	return &result, err
}
