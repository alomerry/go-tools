// Package sdk provides a client for the Apollo Open API.
//
// Usage:
//
//	client := sdk.NewClient("http://localhost:8070", "token")
//	apps, err := client.Apps.GetApps(context.Background(), nil)
//
package sdk
