package git

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubFetch 替换 Fetch 桩，返回还原函数。
func stubFetch(t *testing.T, code int, body string, err error) func() {
	t.Helper()
	old := Fetch
	Fetch = func(_ context.Context, _, _ string) (int, []byte, error) {
		return code, []byte(body), err
	}
	return func() { Fetch = old }
}

func TestBuildURL(t *testing.T) {
	q := Query{Host: "api.github.com", Owner: "o", Repo: "r", Ref: "master", Path: "a/b.md", Page: 0, PageSize: 0}
	got := BuildURL("https://{host}/repos/{owner}/{repo}/contents/{path}?ref={ref}&page={page}&per_page={page_size}", q)
	assert.Equal(t, "https://api.github.com/repos/o/r/contents/a/b.md?ref=master&page=1&per_page=20", got)

	// page_size 上限 100。
	q.Page, q.PageSize = 3, 500
	got = BuildURL("{page}/{page_size}", q)
	assert.Equal(t, "3/100", got)
}

func TestGetFileGithub(t *testing.T) {
	content := base64.StdEncoding.EncodeToString([]byte("# hi\nline2"))
	defer stubFetch(t, 200, `{"type":"file","encoding":"base64","size":12,"sha":"abc123","content":"`+content+`"}`, nil)()
	f, err := GetFile(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r", Path: "README.md"}, "https://h/{owner}/{repo}/{path}")
	require.NoError(t, err)
	assert.Equal(t, "# hi\nline2", f.Content)
	assert.Equal(t, "abc123", f.Sha)
	assert.False(t, f.Truncated)
}

func TestGetFileGithubDir(t *testing.T) {
	defer stubFetch(t, 200, `{"type":"dir"}`, nil)()
	_, err := GetFile(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r", Path: "docs"}, "https://h/x")
	require.ErrorContains(t, err, "directory")
}

func TestGetFileGiteeDirArray(t *testing.T) {
	defer stubFetch(t, 200, `[{"type":"dir","path":"docs"}]`, nil)()
	_, err := GetFile(context.Background(), Query{Host: "example.com", Provider: "gitee.com", Owner: "o", Repo: "r", Path: "docs"}, "https://h/x")
	require.ErrorContains(t, err, "directory")
}

func TestGetFileBinary(t *testing.T) {
	defer stubFetch(t, 200, `{"type":"file","encoding":"none","size":1024,"sha":"deadbeef"}`, nil)()
	f, err := GetFile(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r", Path: "a.png"}, "https://h/x")
	require.NoError(t, err)
	assert.Equal(t, "", f.Content)
	assert.True(t, f.Truncated)
}

func TestGetFileHTTPError(t *testing.T) {
	defer stubFetch(t, 403, `{"message":"rate limit"}`, nil)()
	_, err := GetFile(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r", Path: "a"}, "https://h/x")
	require.ErrorContains(t, err, "http 403")
}

func TestGetFileNetworkError(t *testing.T) {
	defer stubFetch(t, 0, "", errors.New("timeout"))()
	_, err := GetFile(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r", Path: "a"}, "https://h/x")
	require.ErrorContains(t, err, "timeout")
}

func TestGetFileBadArgs(t *testing.T) {
	ctx := context.Background()
	_, err := GetFile(ctx, Query{Host: "example.com", Provider: "gitlab.com", Owner: "o", Repo: "r", Path: "a"}, "https://h/x")
	require.ErrorContains(t, err, "unsupported")

	_, err = GetFile(ctx, Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r", Path: "a"}, "")
	require.ErrorContains(t, err, "no path template")

	_, err = GetFile(ctx, Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r"}, "https://h/x")
	require.ErrorContains(t, err, "required")
}

func TestGetTreeSorts(t *testing.T) {
	// 目录（dir→tree）在前，同组内按 path 字典序；空 sha/path 条目跳过；
	// 上游 size=null 解码为零值。
	defer stubFetch(t, 200, `[{"path":"b.go","type":"file","sha":"2","size":3},{"path":"dir","type":"dir","sha":"4","size":null},{"path":"a.go","type":"file","sha":"1","size":2},{"path":"","type":"file","sha":""}]`, nil)()
	entries, err := GetTree(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r"}, "https://h/x")
	require.NoError(t, err)
	require.Len(t, entries, 3)
	assert.Equal(t, "dir", entries[0].Path)
	assert.Equal(t, "tree", entries[0].Type)
	assert.Equal(t, "a.go", entries[1].Path)
	assert.Equal(t, "blob", entries[1].Type)
	assert.Equal(t, "b.go", entries[2].Path)
}

func TestGetTreeTypeMapping(t *testing.T) {
	// contents type → git/trees 惯例：dir→tree、file/symlink→blob、
	// submodule→commit、未知类型原样透传；path 取上游完整路径。
	defer stubFetch(t, 200, `[
		{"path":"docs","type":"dir","sha":"1","size":0},
		{"path":"README.md","type":"file","sha":"2","size":10},
		{"path":"lnk","type":"symlink","sha":"3","size":5},
		{"path":"sub","type":"submodule","sha":"4","size":0},
		{"path":"weird","type":"fancy","sha":"5","size":0}
	]`, nil)()
	entries, err := GetTree(context.Background(), Query{Host: "example.com", Provider: "gitee.com", Owner: "o", Repo: "r"}, "https://h/x")
	require.NoError(t, err)
	require.Len(t, entries, 5)
	// 排序位次由 TestGetTreeSorts 单独钉；此处按 path 查找只断言类型映射，
	// 不重复钉排序（同一逻辑两处断言，改排序需改两处）。
	byPath := make(map[string]string, len(entries))
	for _, e := range entries {
		byPath[e.Path] = e.Type
	}
	assert.Equal(t, "tree", byPath["docs"])
	assert.Equal(t, "blob", byPath["README.md"])
	assert.Equal(t, "blob", byPath["lnk"])
	assert.Equal(t, "commit", byPath["sub"])
	assert.Equal(t, "fancy", byPath["weird"])
}

func TestGetTreeNonArrayBody(t *testing.T) {
	// 上游限流 403 已被 HTTP 状态校验拦截；此处覆盖 200 但形状不符：把文件
	// 路径传给 tree 模板时上游返回对象（与 parseFile 的数组前置判定对称）。
	defer stubFetch(t, 200, `{"message":"oops"}`, nil)()
	_, err := GetTree(context.Background(), Query{Host: "example.com", Provider: "gitee.com", Owner: "o", Repo: "r", Path: "a.md"}, "https://h/x")
	require.ErrorContains(t, err, "not a directory")
	require.ErrorContains(t, err, "use file api")
}

func TestGetTreeEmptyDir(t *testing.T) {
	// 空目录/（gitee 对不存在路径也）返回空数组：不报错，返回空条目。
	defer stubFetch(t, 200, `[]`, nil)()
	entries, err := GetTree(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r"}, "https://h/x")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestBuildURLTreeRootPath(t *testing.T) {
	// 根目录空 path：{path} 填空后模板自带尾斜杠，产物 ".../contents/?ref=master"。
	// curl 实测（2026-10-20）github/gitee 对带尾斜杠与不带尾斜杠两种形态均
	// 返回 200 数组，无需 BuildURL 特判；本用例钉死产物形态防回归。
	q := Query{Host: "api.github.com", Owner: "o", Repo: "r", Ref: "master", Path: ""}
	got := BuildURL("https://{host}/repos/{owner}/{repo}/contents/{path}?ref={ref}", q)
	assert.Equal(t, "https://api.github.com/repos/o/r/contents/?ref=master", got)

	// 子目录 path 原样拼接（完整路径，非递归语义由上游 contents API 承担）。
	q.Path = "components/git"
	got = BuildURL("https://{host}/repos/{owner}/{repo}/contents/{path}?ref={ref}", q)
	assert.Equal(t, "https://api.github.com/repos/o/r/contents/components/git?ref=master", got)
}

func TestGetHistoryGithub(t *testing.T) {
	defer stubFetch(t, 200, `[{"sha":"abc","commit":{"message":"feat: x\n\ndetail","author":{"name":"amy","date":"2026-10-01T01:00:00Z"}}},{"sha":"","commit":{"message":"bad","author":{"name":"x","date":""}}}]`, nil)()
	commits, err := GetHistory(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r"}, "https://h/x")
	require.NoError(t, err)
	// 空 sha 条目跳过。
	require.Len(t, commits, 1)
	assert.Equal(t, "abc", commits[0].Sha)
	assert.Equal(t, "feat: x\n\ndetail", commits[0].Message)
	assert.Equal(t, "amy", commits[0].Author)
}

func TestListBranches(t *testing.T) {
	defer stubFetch(t, 200, `[{"name":"master","commit":{"sha":"1"}},{"name":"dev","commit":{"sha":"2"}}]`, nil)()
	branches, err := ListBranches(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r"}, "https://h/x")
	require.NoError(t, err)
	require.Len(t, branches, 2)
	assert.Equal(t, "master", branches[0].Name)
}

func TestHistoryPageDefaultsInURL(t *testing.T) {
	var gotURL, gotToken string
	old := Fetch
	Fetch = func(_ context.Context, url, token string) (int, []byte, error) {
		gotURL, gotToken = url, token
		return 200, []byte(`[]`), nil
	}
	defer func() { Fetch = old }()

	_, err := GetHistory(context.Background(), Query{Provider: "gitee.com", Host: "gitee.com", Token: "tok", Owner: "o", Repo: "r", Ref: "master"}, "https://{host}/o/r?page={page}&size={page_size}")
	require.NoError(t, err)
	assert.True(t, strings.Contains(gotURL, "page=1"), gotURL)
	assert.True(t, strings.Contains(gotURL, "size=20"), gotURL)
	assert.Equal(t, "tok", gotToken)
}

func TestBuildURLEscapesSegments(t *testing.T) {
	// Q10：path 按段转义（段间 `/` 保留，段内空格/# 编码）；owner/repo/ref
	// 单段转义；host 不转义。
	q := Query{Host: "api.github.com", Owner: "my org", Repo: "r+e", Ref: "v1.0", Path: "my dir/f#1.md"}
	got := BuildURL("https://{host}/{owner}/{repo}/{path}?ref={ref}", q)
	assert.Equal(t, "https://api.github.com/my%20org/r+e/my%20dir/f%231.md?ref=v1.0", got)

	// 空 path 不 panic、不多 `/`。
	q = Query{Host: "h", Owner: "o", Repo: "r", Ref: "master", Path: ""}
	assert.Equal(t, "https://h/o/r/?ref=master", BuildURL("https://{host}/{owner}/{repo}/{path}?ref={ref}", q))
}

func TestDoFetchHostEmptyFailClosed(t *testing.T) {
	// Q11/S2：host 为空直接失败，不构造残缺 URL 发请求（Fetch 不应被调用）。
	called := false
	old := Fetch
	Fetch = func(_ context.Context, _, _ string) (int, []byte, error) {
		called = true
		return 200, []byte(`{}`), nil
	}
	defer func() { Fetch = old }()

	_, err := GetFile(context.Background(), Query{Provider: "github.com", Owner: "o", Repo: "r", Path: "a"}, "https://h/x")
	require.ErrorContains(t, err, "host is not configured")
	assert.False(t, called)
}

func TestDoFetchBodyTooLarge(t *testing.T) {
	// S5：响应体 >2MB 直接失败，不解析。
	big := make([]byte, (2<<20)+1)
	for i := range big {
		big[i] = 'x'
	}
	defer stubFetch(t, 200, string(big), nil)()
	_, err := GetTree(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r"}, "https://h/x")
	require.ErrorContains(t, err, "too large")
}

func TestUpstreamErrorMapping(t *testing.T) {
	// S4：非 2xx 为结构化 *UpstreamError，状态码可提取；网络错返回 0。
	defer stubFetch(t, 403, "{\"message\":\"rate limit\"}\nretry later", nil)()
	_, err := GetFile(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r", Path: "a"}, "https://h/x")
	require.Error(t, err)
	assert.Equal(t, 403, UpstreamStatusCode(err))
	// S6：snippet 已单行化，错误文案不含换行。
	assert.NotContains(t, err.Error(), "\n")
	assert.NotContains(t, err.Error(), "\r")

	defer stubFetch(t, 0, "", errors.New("timeout"))()
	_, err = GetFile(context.Background(), Query{Host: "example.com", Provider: "github.com", Owner: "o", Repo: "r", Path: "a"}, "https://h/x")
	require.ErrorContains(t, err, "timeout")
	assert.Equal(t, 0, UpstreamStatusCode(err))
}
