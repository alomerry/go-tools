package sdk

import (
	"context"
	"fmt"
	"net/http"
)

type ReleasesService struct {
	client *Client
}

type Release struct {
	AppID                      string            `json:"appId"`
	ClusterName                string            `json:"clusterName"`
	NamespaceName              string            `json:"namespaceName"`
	Name                       string            `json:"name"`
	Configurations             map[string]string `json:"configurations"`
	Comment                    string            `json:"comment"`
	DataChangeCreatedBy        string            `json:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string            `json:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string            `json:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string            `json:"dataChangeLastModifiedTime"`
}

type PublishOptions struct {
	ReleaseTitle   string `json:"releaseTitle"`
	ReleaseComment string `json:"releaseComment"`
	ReleasedBy     string `json:"releasedBy"`
}

func (s *ReleasesService) Publish(ctx context.Context, env, appID, clusterName, namespaceName string, opts *PublishOptions) (*Release, error) {
	path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/releases", env, appID, clusterName, namespaceName)
	var result Release
	_, err := s.client.Do(ctx, http.MethodPost, path, opts, &result)
	return &result, err
}

func (s *ReleasesService) GetLatest(ctx context.Context, env, appID, clusterName, namespaceName string) (*Release, error) {
	path := fmt.Sprintf("/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/releases/latest", env, appID, clusterName, namespaceName)
	var result Release
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return &result, err
}

func (s *ReleasesService) Rollback(ctx context.Context, env string, releaseID int, operator string) error {
	path := fmt.Sprintf("/openapi/v1/envs/%s/releases/%d/rollback?operator=%s", env, releaseID, operator)
	_, err := s.client.Do(ctx, http.MethodPut, path, nil, nil)
	return err
}
