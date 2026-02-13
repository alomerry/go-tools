package sdk

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type AppsService struct {
	client *Client
}

type AppEnvCluster struct {
	Env      string   `json:"env"`
	Clusters []string `json:"clusters"`
}

type App struct {
	Name                       string `json:"name"`
	AppID                      string `json:"appId"`
	OrgID                      string `json:"orgId"`
	OrgName                    string `json:"orgName"`
	OwnerName                  string `json:"ownerName"`
	OwnerEmail                 string `json:"ownerEmail"`
	DataChangeCreatedBy        string `json:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string `json:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string `json:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string `json:"dataChangeLastModifiedTime"`
}

// GetEnvClusters returns the environment and cluster information for an app.
// URL: http://{portal_address}/openapi/v1/apps/{appId}/envclusters
func (s *AppsService) GetEnvClusters(ctx context.Context, appID string) ([]*AppEnvCluster, error) {
	path := fmt.Sprintf("/openapi/v1/apps/%s/envclusters", appID)
	var result []*AppEnvCluster
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return result, err
}

// GetApps returns the information for all apps or specific apps.
// URL: http://{portal_address}/openapi/v1/apps
// appIDs: optional, comma separated appIds
func (s *AppsService) GetApps(ctx context.Context, appIDs []string) ([]*App, error) {
	path := "/openapi/v1/apps"
	if len(appIDs) > 0 {
		path += fmt.Sprintf("?appIds=%s", strings.Join(appIDs, ","))
	}
	var result []*App
	_, err := s.client.Do(ctx, http.MethodGet, path, nil, &result)
	return result, err
}
