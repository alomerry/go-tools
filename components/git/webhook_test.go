package git

import (
	"context"
	"errors"
	stdhttp "net/http"
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

// stubFetchHeaders 替换 FetchHeaders 桩（deliveries 游标分页取数用），返回
// 还原函数与请求 URL 捕获。
func stubFetchHeaders(t *testing.T, code int, link, body string, err error) (func(), *string) {
	t.Helper()
	var gotURL string
	old := FetchHeaders
	FetchHeaders = func(_ context.Context, url, _ string) (int, stdhttp.Header, []byte, error) {
		gotURL = url
		h := make(stdhttp.Header)
		if link != "" {
			h.Set("Link", link)
		}
		return code, h, []byte(body), err
	}
	return func() { FetchHeaders = old }, &gotURL
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
	restore, _ := stubFetchHeaders(t, 200, "", body, nil)
	defer restore()
	ds, next, err := ListDeliveries(context.Background(), Query{Provider: ProviderGithub, Host: "api.github.com", Owner: "o", Repo: "r", HookID: 3}, "https://{host}/repos/{owner}/{repo}/hooks/{hook_id}/deliveries?per_page={page_size}")
	require.NoError(t, err)
	require.Len(t, ds, 1)
	assert.Equal(t, int64(7), ds[0].ID)
	assert.Equal(t, "push", ds[0].Event)
	assert.Equal(t, "published", ds[0].Action)
	assert.Equal(t, "OK", ds[0].Status)
	assert.Equal(t, 200, ds[0].StatusCode)
	assert.False(t, ds[0].Redelivered)
	// 无 Link 头（单页）→ next 为空。
	assert.Empty(t, next)
}

func TestListDeliveriesCursorRoundTrip(t *testing.T) {
	body := `[{"id":7,"delivered_at":"2026-06-03T00:57:16Z","redelivered":false,"status":"OK","status_code":200,"event":"push"}]`
	link := `<https://api.github.com/repos/o/r/hooks/3/deliveries?per_page=1&cursor=CURSOR2>; rel="next", <https://api.github.com/repos/o/r/hooks/3/deliveries?per_page=1>; rel="prev"`
	restore, gotURL := stubFetchHeaders(t, 200, link, body, nil)
	defer restore()
	ds, next, err := ListDeliveries(context.Background(), Query{Provider: ProviderGithub, Host: "api.github.com", Owner: "o", Repo: "r", HookID: 3, Cursor: "CURSOR1", PageSize: 1}, "https://{host}/repos/{owner}/{repo}/hooks/{hook_id}/deliveries?per_page={page_size}")
	require.NoError(t, err)
	require.Len(t, ds, 1)
	assert.Equal(t, "CURSOR2", next)
	// 发往上游的 URL：cursor 透传、per_page 归一、无 page 键（传 page 即 422）。
	u, err := stdhttp.NewRequest("GET", *gotURL, nil)
	require.NoError(t, err)
	query := u.URL.Query()
	assert.Equal(t, "CURSOR1", query.Get("cursor"))
	assert.Equal(t, "1", query.Get("per_page"))
	assert.Empty(t, query.Get("page"))
}

func TestListDeliveriesTemplateWithPageFailFast(t *testing.T) {
	called := false
	old := FetchHeaders
	FetchHeaders = func(_ context.Context, _, _ string) (int, stdhttp.Header, []byte, error) {
		called = true
		return 200, nil, []byte(`[]`), nil
	}
	defer func() { FetchHeaders = old }()

	// 旧 Apollo 模板残留 {page}：上线前 fail-fast，不发请求（发即上游 422）。
	_, _, err := ListDeliveries(context.Background(), Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r", HookID: 3}, "https://{host}/repos/{owner}/{repo}/hooks/{hook_id}/deliveries?page={page}&per_page={page_size}")
	require.ErrorContains(t, err, "{page}")
	assert.False(t, called)

	// {page_size} 含 page 子串但无花括号闭合：不得误判，照常填充。
	_, _, err = ListDeliveries(context.Background(), Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r", HookID: 3}, "https://{host}/x?per_page={page_size}")
	require.NoError(t, err)
	assert.True(t, called)
}

func TestBuildDeliveriesURL(t *testing.T) {
	// 空游标（首页）：不发送 cursor 参数，非空经 query 统一装配。
	got, err := BuildDeliveriesURL("https://{host}/repos/{owner}/{repo}/hooks/{hook_id}/deliveries?per_page={page_size}", Query{Provider: ProviderGithub, Host: "api.github.com", Owner: "o", Repo: "r", HookID: 3})
	require.NoError(t, err)
	assert.Equal(t, "https://api.github.com/repos/o/r/hooks/3/deliveries?per_page=20", got)

	// 非空游标经 query 统一装配（模板无需 {cursor} 占位）。
	got, err = BuildDeliveriesURL("https://{host}/x?per_page={page_size}", Query{Host: "h", Cursor: "ab cd"})
	require.NoError(t, err)
	assert.Contains(t, got, "cursor=ab+cd")

	// 模板 query 位含 {cursor} 占位同样收敛到游标值（query.Set 覆盖，
	// 非 replacer 填充；空游标则删参，不残留字面量）。
	got, err = BuildDeliveriesURL("https://{host}/x?cursor={cursor}&per_page={page_size}", Query{Host: "h", Cursor: "NEXT1"})
	require.NoError(t, err)
	assert.Contains(t, got, "cursor=NEXT1")
	assert.NotContains(t, got, "{cursor}")
	got, err = BuildDeliveriesURL("https://{host}/x?cursor={cursor}&per_page={page_size}", Query{Host: "h"})
	require.NoError(t, err)
	assert.NotContains(t, got, "cursor")

	// pageSize 归一（缺省 20/上限 100，与 NormalizedPage 同口径）。
	got, err = BuildDeliveriesURL("https://h/x?per_page={page_size}", Query{PageSize: 500})
	require.NoError(t, err)
	assert.Contains(t, got, "per_page=100")

	// 空模板 fail-closed。
	_, err = BuildDeliveriesURL("", Query{Provider: ProviderGithub})
	require.ErrorContains(t, err, "no path template")
}

func TestParseNextCursor(t *testing.T) {
	link := `<https://api.github.com/repos/o/r/hooks/3/deliveries?per_page=1&cursor=NEXT1>; rel="next", <https://api.github.com/repos/o/r/hooks/3/deliveries?per_page=1>; rel="prev"`
	assert.Equal(t, "NEXT1", parseNextCursor(link))
	// 仅 prev（首页无 next 或末页）→ 空。
	assert.Empty(t, parseNextCursor(`<https://api.github.com/x?cursor=P>; rel="prev"`))
	// 无 Link 头（单页）→ 空。
	assert.Empty(t, parseNextCursor(""))
	// 畸形段容忍：坏段跳过，好段照常提取。
	assert.Equal(t, "N2", parseNextCursor(`not-a-link, <https://h/x?cursor=N2>; rel="next"`))
	// url 编码 cursor 正确解码。
	assert.Equal(t, "a/b+c", parseNextCursor(`<https://h/x?cursor=a%2Fb%2Bc>; rel="next"`))
}

func TestListDeliveriesGiteeUnsupported(t *testing.T) {
	called := false
	old := FetchHeaders
	FetchHeaders = func(_ context.Context, _, _ string) (int, stdhttp.Header, []byte, error) {
		called = true
		return 200, nil, []byte(`[]`), nil
	}
	defer func() { FetchHeaders = old }()

	// gitee v5 无 deliveries API：即便误配了模板，provider 白名单直接拒绝，不发请求。
	_, _, err := ListDeliveries(context.Background(), Query{Provider: ProviderGitee, Host: "gitee.com", Owner: "o", Repo: "r", HookID: 3}, "https://{host}/api/v5/x")
	require.ErrorContains(t, err, "does not support webhook deliveries")
	assert.False(t, called)

	// hook_id 缺失同样拒绝。
	_, _, err = ListDeliveries(context.Background(), Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r"}, "https://h/x")
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

	// 非 2xx 且响应体超 2MB：snippet 以 "response too large" 固定文案替代
	//（与 doFetch 纵深防御同口径），StatusCode 保留供按码映射。
	defer stubPostFetch(t, 500, strings.Repeat("x", maxBodyBytes+1), nil)()
	err = RedeliverDelivery(context.Background(), Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r", HookID: 3, DeliveryID: 7}, "https://h/x")
	require.Error(t, err)
	assert.Equal(t, 500, UpstreamStatusCode(err))
	assert.Contains(t, err.Error(), "response too large")
	assert.NotContains(t, err.Error(), "\n")

	// 边界：非 2xx 响应体恰为 2MB 不超限，snippet 原样保留。
	defer stubPostFetch(t, 500, strings.Repeat("x", maxBodyBytes), nil)()
	err = RedeliverDelivery(context.Background(), Query{Provider: ProviderGithub, Host: "h", Owner: "o", Repo: "r", HookID: 3, DeliveryID: 7}, "https://h/x")
	require.Error(t, err)
	assert.Equal(t, 500, UpstreamStatusCode(err))
	assert.NotContains(t, err.Error(), "response too large")

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
