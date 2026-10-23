package git

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubPostFetch 替换 PostFetch 桩，返回还原函数（契约同 stubFetch）。
func stubPostFetch(t *testing.T, code int, body string, err error) func() {
	t.Helper()
	old := PostFetch
	PostFetch = func(_ context.Context, _, _ string) (int, []byte, error) {
		return code, []byte(body), err
	}
	return func() { PostFetch = old }
}

func TestListHooksGithub(t *testing.T) {
	body := `[{"id":1,"name":"web","active":true,"config":{"url":"https://example.com/hook","content_type":"json"},"events":["push","pull_request"],"created_at":"2026-01-02T03:04:05Z"},{"id":0,"name":"bad"}]`
	defer stubFetch(t, 200, body, nil)()
	hooks, err := ListHooks(context.Background(), Query{Provider: ProviderGithub, Host: "api.github.com", Owner: "o", Repo: "r"}, "https://{host}/repos/{owner}/{repo}/hooks")
	require.NoError(t, err)
	require.Len(t, hooks, 1)
	assert.Equal(t, int64(1), hooks[0].ID)
	assert.Equal(t, "https://example.com/hook", hooks[0].URL)
	assert.True(t, hooks[0].Active)
	assert.Equal(t, []string{"push", "pull_request"}, hooks[0].Events)
}

func TestListHooksGitee(t *testing.T) {
	body := `[{"id":42,"url":"https://example.com/gitee-hook","created_at":"2026-04-16T17:35:56+08:00","password":"x","push_events":true,"tag_push_events":false,"issues_events":true,"note_events":false,"project_id":1}]`
	defer stubFetch(t, 200, body, nil)()
	hooks, err := ListHooks(context.Background(), Query{Provider: ProviderGitee, Host: "gitee.com", Owner: "o", Repo: "r"}, "https://{host}/api/v5/repos/{owner}/{repo}/hooks")
	require.NoError(t, err)
	require.Len(t, hooks, 1)
	assert.Equal(t, int64(42), hooks[0].ID)
	assert.Equal(t, "https://example.com/gitee-hook", hooks[0].URL)
	// gitee 无停用概念恒生效；事件由布尔位派生，false 位不出现在列表。
	assert.True(t, hooks[0].Active)
	assert.Equal(t, []string{"push", "issues"}, hooks[0].Events)
}

func TestListHooksBadArgs(t *testing.T) {
	defer stubFetch(t, 200, `[]`, nil)()
	_, err := ListHooks(context.Background(), Query{Provider: ProviderGithub, Host: "h", Repo: "r"}, "https://h/x")
	require.ErrorContains(t, err, "owner/repo are required")
}

func TestListDeliveriesGithub(t *testing.T) {
	body := `[{"id":7,"guid":"g-1","delivered_at":"2026-06-03T00:57:16Z","redelivered":false,"duration":0.27,"status":"OK","status_code":200,"event":"push","action":"published","repository_id":2},{"id":0,"guid":"bad"}]`
	defer stubFetch(t, 200, body, nil)()
	ds, err := ListDeliveries(context.Background(), Query{Provider: ProviderGithub, Host: "api.github.com", Owner: "o", Repo: "r", HookID: 3}, "https://{host}/repos/{owner}/{repo}/hooks/{hook_id}/deliveries")
	require.NoError(t, err)
	require.Len(t, ds, 1)
	assert.Equal(t, int64(7), ds[0].ID)
	assert.Equal(t, "push", ds[0].Event)
	assert.Equal(t, "published", ds[0].Action)
	assert.Equal(t, "OK", ds[0].Status)
	assert.Equal(t, 200, ds[0].StatusCode)
	assert.False(t, ds[0].Redelivered)
}

func TestListDeliveriesGiteeUnsupported(t *testing.T) {
	called := false
	old := Fetch
	Fetch = func(_ context.Context, _, _ string) (int, []byte, error) {
		called = true
		return 200, []byte(`[]`), nil
	}
	defer func() { Fetch = old }()

	// gitee v5 无 deliveries API：即便误配了模板，provider 白名单直接拒绝，不发请求。
	_, err := ListDeliveries(context.Background(), Query{Provider: ProviderGitee, Host: "gitee.com", Owner: "o", Repo: "r", HookID: 3}, "https://{host}/api/v5/x")
	require.ErrorContains(t, err, "does not support webhook deliveries")
	assert.False(t, called)

	// hook_id 缺失同样拒绝。
	_, err = ListDeliveries(context.Background(), Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r"}, "https://h/x")
	require.ErrorContains(t, err, "hook_id is required")

	// 能力判定与白名单同源：github true，其余 false。
	assert.True(t, SupportsDeliveries(ProviderGithub))
	assert.False(t, SupportsDeliveries(ProviderGitee))
	assert.False(t, SupportsDeliveries("gitlab.com"))
}

func TestRedeliverDelivery(t *testing.T) {
	var gotURL, gotToken string
	old := PostFetch
	PostFetch = func(_ context.Context, url, token string) (int, []byte, error) {
		gotURL, gotToken = url, token
		return 204, nil, nil
	}
	defer func() { PostFetch = old }()

	// 上游 204 无 body 即成功。
	require.NoError(t, RedeliverDelivery(context.Background(), Query{Provider: ProviderGithub, Host: "api.github.com", Token: "tk", Owner: "o", Repo: "r", HookID: 3, DeliveryID: 7}, "https://{host}/repos/{owner}/{repo}/hooks/{hook_id}/deliveries/{delivery_id}/attempts"))
	assert.Equal(t, "https://api.github.com/repos/o/r/hooks/3/deliveries/7/attempts", gotURL)
	assert.Equal(t, "tk", gotToken)

	// 非 2xx → 结构化 UpstreamError。
	defer stubPostFetch(t, 403, `{"message":"Not Found"}`, nil)()
	err := RedeliverDelivery(context.Background(), Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r", HookID: 3, DeliveryID: 7}, "https://h/x")
	require.Error(t, err)
	assert.Equal(t, 403, UpstreamStatusCode(err))
	assert.NotContains(t, err.Error(), "\n")

	// 网络错透传。
	defer stubPostFetch(t, 0, "", errors.New("dial timeout"))()
	err = RedeliverDelivery(context.Background(), Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r", HookID: 3, DeliveryID: 7}, "https://h/x")
	require.ErrorContains(t, err, "dial timeout")
}

func TestRedeliverDeliveryBadArgs(t *testing.T) {
	called := false
	old := PostFetch
	PostFetch = func(_ context.Context, _, _ string) (int, []byte, error) {
		called = true
		return 204, nil, nil
	}
	defer func() { PostFetch = old }()

	q := Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r", HookID: 3}
	require.ErrorContains(t, RedeliverDelivery(context.Background(), q, "https://h/x"), "delivery_id")
	assert.False(t, called)

	// gitee 白名单拒绝。
	q.Provider, q.DeliveryID = ProviderGitee, 7
	require.ErrorContains(t, RedeliverDelivery(context.Background(), q, "https://h/x"), "does not support webhook deliveries")
	assert.False(t, called)

	// 空模板 fail-closed。
	q.Provider = ProviderGithub
	require.ErrorContains(t, RedeliverDelivery(context.Background(), q, ""), "no path template")
	assert.False(t, called)
}

func TestBuildURLHookPlaceholders(t *testing.T) {
	got, err := BuildURL("https://{host}/repos/{owner}/{repo}/hooks/{hook_id}/deliveries/{delivery_id}/attempts", Query{Host: "api.github.com", Owner: "o", Repo: "r", HookID: 3, DeliveryID: 7})
	require.NoError(t, err)
	assert.Equal(t, "https://api.github.com/repos/o/r/hooks/3/deliveries/7/attempts", got)
}

func TestListHooksUpstreamError(t *testing.T) {
	defer stubFetch(t, 401, `{"message":"Bad credentials"}`, nil)()
	_, err := ListHooks(context.Background(), Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r"}, "https://h/x")
	require.Error(t, err)
	assert.Equal(t, 401, UpstreamStatusCode(err))
	assert.False(t, strings.Contains(err.Error(), "\n"))
}
