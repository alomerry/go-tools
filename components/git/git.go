// Package git 是 git 页三段式只读浏览的统一 SDK（service-git-page）。
//
// 对外四个浏览能力：GetFile（单文件内容）、GetTree（目录直接子项，非递归）、
// GetHistory（提交历史分页）、ListBranches（分支列表）——github/gitee
// 两家返回形状同形（message/author 嵌套在 commit 下），解析共用一套
// struct+parse，不再按 provider 双实现。token 与 host 由调用方从 apollo
// GitCfg().Providers 取出后传入——本包不直接依赖 apollo，方便单测与复用。
//
// webhook 能力（service-git-webhook，webhook.go）：ListHooks（hooks 配置
// 列表，两家形状不同、按 provider 内部分派解析出同形 Hook）、
// ListDeliveries（github 调用记录分页）、RedeliverDelivery（github 单条
// 重发，POST）。后两能力 gitee v5 无对应 API，显式 provider 白名单。
//
// URL 模板来自 GitItemConfig 的路径模板（浏览四件 + hooksPath/
// deliveriesPath/redeliverPath），占位符
// {host}/{owner}/{repo}/{ref}/{path}/{commit}/{page}/{page_size}/
// {hook_id}/{delivery_id} 由 BuildURL 填充。空模板 = 该 provider 不支持
// 对应能力，直接返回明确错误。
package git

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	stdhttp "net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/alomerry/go-tools/components/http"
	"github.com/alomerry/go-tools/components/http/opts/req"
)

const (
	ProviderGithub = "github.com"
	ProviderGitee  = "gitee.com"

	// DefaultPageSize 历史分页缺省与上限：与后端 git_browse.go 同口径。
	DefaultPageSize = 20
	MaxPageSize     = 100

	// maxBodyBytes 上游单次响应体上限 2MB（S5）：超过直接失败，不解析不缓存，
	// 防超大文件/错误页撑内存。文件内容截断（100KB/2000 行）是展示口径，
	// 在此之前——2MB 是传输安全闸。
	maxBodyBytes = 2 << 20
)

// UpstreamError 上游非 2xx 的结构化错误（S4）：携带 HTTP 状态码供调用方
// （homelab-backend gitBrowseUnavailable）按码映射 gRPC 码；Snippet 已做
// 单行化（\n\r→空格）与 200 字节截断，只进日志不进响应。
type UpstreamError struct {
	StatusCode int
	Snippet    string
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("git browse failed: http %d: %s", e.StatusCode, e.Snippet)
}

// UpstreamStatusCode 从 err 中提取上游 HTTP 状态码（S4 映射用）：
// 非 2xx 的 *UpstreamError 返回其 StatusCode；其余（网络错/解析错/参数错）
// 返回 0，调用方一律按 Unavailable 处理。
func UpstreamStatusCode(err error) int {
	var ue *UpstreamError
	if err != nil && errors.As(err, &ue) {
		return ue.StatusCode
	}
	return 0
}

// Query 单次浏览查询的全部入参（host/token 由调用方从 apollo Providers 取）。
type Query struct {
	Provider string
	Host     string
	Token    string
	Owner    string
	Repo     string
	// Ref 分支或 commit；空由调用方填缺省（后端缺省 master）。
	Ref string
	// Path 文件或目录路径（仓库相对路径）。
	Path string
	// Commit 历史过滤（可选，指定 commit 附近的记录）。
	Commit string
	Page   int
	// HookID / DeliveryID webhook 能力占位（ListDeliveries/RedeliverDelivery 用），
	// 浏览四能力不消费。
	HookID     int64
	DeliveryID int64
	// Cursor github deliveries 游标分页（ListDeliveries 用）：空即首页，非空为
	// 上一页响应 Link 头 rel="next" 的 cursor。浏览四能力与 ListHooks 不消费。
	Cursor string
	// PageSize 分页大小。
	PageSize int
}

// FileContent 单文件内容：Content 已做 base64 解码（encoding=="none" 时为
// 二进制/超大文件，Content 为空，调用方仅提示不渲染）。
type FileContent struct {
	Path      string
	Content   string
	Encoding  string
	Sha       string
	Size      int64
	Truncated bool
}

// TreeEntry 树节点（目录单层直接子项）：Type 为 "tree"（目录）、"blob"
// （文件/符号链接）或 "commit"（子模块），对齐 git/trees 惯例；未知类型
// 原样透传。Path 为仓库相对完整路径（如 "lib/main.dart"）。
type TreeEntry struct {
	Path string
	Type string
	Sha  string
	Size int64
}

// Commit 单条提交：Message 为完整信息，首行截取由展示层负责。
type Commit struct {
	Sha     string
	Message string
	Author  string
	Date    string
}

// Branch 分支：仅名称，默认分支判定由调用方负责。
type Branch struct {
	Name string
}

// FetchFunc 单次 HTTP 取数（statusCode + body），变量化以便单测桩替换。
type FetchFunc func(ctx context.Context, url, token string) (statusCode int, body []byte, err error)

// Fetch 默认取数实现：FetchHeaders 的一行 wrapper（共享 resty client +
// token 鉴权内核，抛弃响应头）。
var Fetch FetchFunc = func(ctx context.Context, url, token string) (int, []byte, error) {
	code, _, body, err := FetchHeaders(ctx, url, token)
	return code, body, err
}

// FetchHeadersFunc 带响应头取数（statusCode + header + body），变量化以便单测
// 桩替换。deliveries 游标分页需读上游 Link 头取下一页 cursor，Fetch 只回
// body 满足不了，走本变量；其余能力沿用 Fetch 不动。
type FetchHeadersFunc func(ctx context.Context, url, token string) (statusCode int, header stdhttp.Header, body []byte, err error)

// FetchHeaders 默认带头取数实现：共享 resty client + token 鉴权。
var FetchHeaders FetchHeadersFunc = func(ctx context.Context, url, token string) (int, stdhttp.Header, []byte, error) {
	resp, err := http.GetClient().Get(ctx, url, req.WithAuthToken(token))
	if err != nil {
		return 0, nil, nil, err
	}
	return resp.StatusCode(), resp.Header(), resp.Bytes(), nil
}

// IsSupportedProvider 是否为 SDK 内建解析的 provider。
func IsSupportedProvider(name string) bool {
	return name == ProviderGithub || name == ProviderGitee
}

// BuildURL 填充路径模板占位符。page/pageSize<=0 时按缺省（1/20）填充，
// pageSize 上限 100——SDK 直调同样受页大小约束，不依赖后端二次收敛。
// 转义（Q10/S2）：path 按段 PathEscape（保留段间 `/`，段内空格等特殊字符
// 编码）；ref/owner/repo/commit 为单段 PathEscape；host 为主机名不转义；
// page/page_size 为归一化整数不转义。path 含 `.`/`..` 段时返回错误（F1），
// 调用方（doFetch 及直调 BuildURL 者）不得忽略。
func BuildURL(tmpl string, q Query) (string, error) {
	page, pageSize := NormalizedPage(q.Page, q.PageSize)
	path, err := escapeGitPath(q.Path)
	if err != nil {
		return "", err
	}
	r := strings.NewReplacer(
		"{host}", q.Host,
		"{owner}", url.PathEscape(q.Owner),
		"{repo}", url.PathEscape(q.Repo),
		"{ref}", url.PathEscape(q.Ref),
		"{path}", path,
		"{commit}", url.PathEscape(q.Commit),
		"{page}", strconv.Itoa(page),
		"{page_size}", strconv.Itoa(pageSize),
		// webhook 能力占位：浏览模板不含这两占位符，填充无副作用。
		"{hook_id}", strconv.FormatInt(q.HookID, 10),
		"{delivery_id}", strconv.FormatInt(q.DeliveryID, 10),
	)
	return r.Replace(tmpl), nil
}

// BuildDeliveriesURL deliveries 专用 URL 装配（github cursor 分页，无 page 语义）：
//   - 空模板 fail-closed（与 doFetch 同口径，防误配模板打到错误平台）。
//   - 模板含 `{page}` 直接 fail-fast：github deliveries 端点以 422 拒绝 page
//     键（InvalidArgument 非法键），静默忽略会造成「翻页无反应」的错数据，
//     必须在部署期暴露而非线上 502。
//   - `{page_size}` 照常填充为 per_page 值（归一缺省 20/上限 100）；游标仅
//     支持 query 形态、经 URL query 统一装配（兼容 Apollo 止血模板
//     `?per_page={page_size}`）：模板 query 位含 `{cursor}` 占位会被游标值覆
//     盖、无占位则自动追加（path 位的 `{cursor}` 不支持，会残留字面量打到上
//     游）；空游标（首页）不发送 cursor 参数。
//
// 注意 `{page}` 判定须精确匹配：`{page_size}` 含 "page" 子串但无花括号闭合，
// 不会误判。
func BuildDeliveriesURL(tmpl string, q Query) (string, error) {
	if tmpl == "" {
		return "", fmt.Errorf("provider %s has no path template for this operation", q.Provider)
	}
	if strings.Contains(tmpl, "{page}") {
		return "", fmt.Errorf("deliveries template must not contain {page}: github deliveries API is cursor-paginated (per_page/cursor), {page} is rejected with http 422")
	}
	_, pageSize := NormalizedPage(0, q.PageSize)
	r := strings.NewReplacer(
		"{host}", q.Host,
		"{owner}", url.PathEscape(q.Owner),
		"{repo}", url.PathEscape(q.Repo),
		"{hook_id}", strconv.FormatInt(q.HookID, 10),
		"{page_size}", strconv.Itoa(pageSize),
	)
	u, err := url.Parse(r.Replace(tmpl))
	if err != nil {
		return "", err
	}
	query := u.Query()
	if strings.TrimSpace(q.Cursor) == "" {
		query.Del("cursor")
	} else {
		query.Set("cursor", strings.TrimSpace(q.Cursor))
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}

// escapeGitPath 仓库相对路径按段转义：段间 `/` 原样保留，段内逐段
// PathEscape（如空格→%20）。空串/首尾斜杠/连续斜杠（空段）形态保持
// （Split/Join 自然语义，不在穿越防护范围）。
// 穿越防护（F1）：`url.PathEscape` 对 `.`/`..` 原样返回，`a/../../x` 原样
// 进 URL 后由上游按 RFC 归一化越出 `/contents/` 前缀，可借服务端 token
// 盲探同 host 任意 GET 端点——任一段为 `.` 或 `..` 即明确拒绝。
func escapeGitPath(p string) (string, error) {
	parts := strings.Split(p, "/")
	for i, s := range parts {
		if s == "." || s == ".." {
			return "", fmt.Errorf("invalid path segment %q", s)
		}
		parts[i] = url.PathEscape(s)
	}
	return strings.Join(parts, "/"), nil
}

// NormalizedPage 分页归一：page/pageSize<=0 时按缺省（1/20）填充，
// pageSize 上限 100——SDK 直调同样受页大小约束，不依赖后端二次收敛。
// 后端回显 page/page_size 与 SDK 同口径直接复用本函数。
func NormalizedPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}

// doFetch 四个 Get* 的公共前置：provider/模板/host 校验后取数 + HTTP
// 状态校验（resty 对非 2xx 默认不报错，上游限流 403 + 错误 JSON 必须在此
// 拦截，避免静默产出空快照）+ 响应体上限（2MB，S5）。
// Fail-closed（Q11/S2）：host 为空（apollo 未配该 provider 的 host）直接
// 返回明确错误，不构造残缺 URL 发请求。
func doFetch(ctx context.Context, q Query, tmpl string) ([]byte, error) {
	if !IsSupportedProvider(q.Provider) {
		return nil, fmt.Errorf("unsupported git provider %s", q.Provider)
	}
	if tmpl == "" {
		return nil, fmt.Errorf("provider %s has no path template for this operation", q.Provider)
	}
	if strings.TrimSpace(q.Host) == "" {
		return nil, fmt.Errorf("git host is not configured for provider %s", q.Provider)
	}
	target, err := BuildURL(tmpl, q)
	if err != nil {
		return nil, err
	}
	code, body, err := Fetch(ctx, target, q.Token)
	if err != nil {
		return nil, err
	}
	if err := checkBrowseBody(code, body); err != nil {
		return nil, err
	}
	return body, nil
}

// newUpstreamError 非 2xx 结构化错误单点（S4）：响应体超 2MB 时以固定文案
// 替代 snippet，防超大错误页整体进错误链路；StatusCode 保留供调用方按码映射。
// 2MB 口径（S5）收敛于此与 checkBrowseBody 的成功路径上限两处，文案格式单点。
func newUpstreamError(code int, body []byte) *UpstreamError {
	if len(body) > maxBodyBytes {
		return &UpstreamError{StatusCode: code, Snippet: fmt.Sprintf("response too large (%d bytes, limit %d)", len(body), maxBodyBytes)}
	}
	return &UpstreamError{StatusCode: code, Snippet: snippet(body)}
}

// checkBrowseBody 取数后置校验单点（2MB 上限 S5 + 非 2xx 结构化 + 空 body
// 拒绝）：doFetch 与 ListDeliveries（走 FetchHeaders 无头内核，无法复用
// doFetch 整体）同调，避免口径漂移。
func checkBrowseBody(code int, body []byte) error {
	if len(body) > maxBodyBytes {
		return fmt.Errorf("git browse failed: response too large (%d bytes, limit %d)", len(body), maxBodyBytes)
	}
	if code/100 != 2 {
		return newUpstreamError(code, body)
	}
	if len(body) == 0 {
		return fmt.Errorf("git browse failed: empty response body")
	}
	return nil
}

// GetFile 取单文件内容（content 已 base64 解码）。
func GetFile(ctx context.Context, q Query, tmpl string) (*FileContent, error) {
	if q.Owner == "" || q.Repo == "" || q.Path == "" {
		return nil, fmt.Errorf("owner/repo/path are required")
	}
	body, err := doFetch(ctx, q, tmpl)
	if err != nil {
		return nil, err
	}
	return parseFile(body, q.Path)
}

// GetTree 取 path 目录的直接子项（单层，非递归；path 为空即仓库根目录）。
// 返回按 "tree"（目录）在前、再按 path 字典序的 deterministic 排序，供
// 树形展示直接渲染（GitHub 同口径）。
func GetTree(ctx context.Context, q Query, tmpl string) ([]TreeEntry, error) {
	if q.Owner == "" || q.Repo == "" {
		return nil, fmt.Errorf("owner/repo are required")
	}
	body, err := doFetch(ctx, q, tmpl)
	if err != nil {
		return nil, err
	}
	entries, err := parseTree(body, q.Path)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		if (entries[i].Type == "tree") != (entries[j].Type == "tree") {
			return entries[i].Type == "tree"
		}
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

// GetHistory 取提交历史（分页由模板的 {page}/{page_size} 透给上游）。
func GetHistory(ctx context.Context, q Query, tmpl string) ([]Commit, error) {
	if q.Owner == "" || q.Repo == "" {
		return nil, fmt.Errorf("owner/repo are required")
	}
	body, err := doFetch(ctx, q, tmpl)
	if err != nil {
		return nil, err
	}
	return parseHistory(body)
}

// ListBranches 列分支。
func ListBranches(ctx context.Context, q Query, tmpl string) ([]Branch, error) {
	if q.Owner == "" || q.Repo == "" {
		return nil, fmt.Errorf("owner/repo are required")
	}
	body, err := doFetch(ctx, q, tmpl)
	if err != nil {
		return nil, err
	}
	return parseBranches(body)
}

func snippet(body []byte) string {
	const max = 200
	singleLine := func(s string) string {
		// S6：\n\r→空格——snippet 会进错误文案/日志，换行不处理即日志注入
		// （多行伪造）与单行断言失效；先折叠再截断。
		s = strings.ReplaceAll(s, "\r", " ")
		s = strings.ReplaceAll(s, "\n", " ")
		return strings.TrimSpace(s)
	}
	if len(body) <= max {
		return singleLine(string(body))
	}
	return singleLine(string(body[:max])) + "..."
}

// --- 上游解析（github/gitee 同形，共用一套） ---

type gitFile struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Size     int64  `json:"size"`
	Sha      string `json:"sha"`
	Content  string `json:"content"`
	// Truncated 大文件 API 标记（超限时 content 为空）。
	Truncated bool `json:"truncated"`
}

func parseFile(body []byte, path string) (*FileContent, error) {
	// gitee 目录查询返回数组而非对象：形状不符即明确报错，提示用 tree 接口。
	if len(body) > 0 && body[0] == '[' {
		return nil, fmt.Errorf("path %s is a directory, use tree api", path)
	}
	var f gitFile
	if err := json.Unmarshal(body, &f); err != nil {
		return nil, fmt.Errorf("parse git file failed: %w", err)
	}
	if f.Type == "dir" {
		return nil, fmt.Errorf("path %s is a directory, use tree api", path)
	}
	// encoding=="none"（二进制/超大）时 content 为空：不报错，调用方仅提示。
	if f.Encoding == "none" {
		return &FileContent{Path: path, Encoding: "none", Sha: f.Sha, Size: f.Size, Truncated: true}, nil
	}
	if f.Content == "" && f.Size > 0 {
		return nil, fmt.Errorf("empty file content for %s", path)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(f.Content, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("decode git file content failed: %w", err)
	}
	return &FileContent{Path: path, Content: string(raw), Encoding: f.Encoding, Sha: f.Sha, Size: f.Size, Truncated: f.Truncated}, nil
}

type gitTreeEntry struct {
	// Type 上游口径："file"/"dir"/"symlink"/"submodule"（github/gitee 同形）。
	Type string `json:"type"`
	// Size 目录条目可能为 null（gitee 实测），null 解码为零值不报错。
	Size int64  `json:"size"`
	Sha  string `json:"sha"`
	Path string `json:"path"`
}

// parseTree 解析 contents API 目录查询返回的 JSON 数组（github/gitee 同形）。
// 类型映射对齐 git/trees 惯例：dir→tree、file→blob、symlink→blob、
// submodule→commit，未知类型原样透传（上游新增类型不丢条目）。
// 与 parseFile 的数组前置判定对称：body 为对象（如把文件路径传给了 tree
// 模板）即明确报错，不静默产出零值快照。
func parseTree(body []byte, path string) ([]TreeEntry, error) {
	if len(body) > 0 && body[0] == '{' {
		return nil, fmt.Errorf("path %s is not a directory, use file api", path)
	}
	var list []gitTreeEntry
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parse git tree failed: %w", err)
	}
	out := make([]TreeEntry, 0, len(list))
	for _, e := range list {
		// 空 sha/path 条目跳过（gitee 限流 403 + 错误 JSON 能过 unmarshal 但无
		// sha，此处兜底；HTTP 状态校验是第一道防线）。
		if e.Sha == "" || e.Path == "" {
			continue
		}
		out = append(out, TreeEntry{Path: e.Path, Type: mapTreeEntryType(e.Type), Sha: e.Sha, Size: e.Size})
	}
	return out, nil
}

// mapTreeEntryType contents API 的 type → git/trees 惯例 type；
// 未知类型原样透传。
func mapTreeEntryType(t string) string {
	switch t {
	case "dir":
		return "tree"
	case "file", "symlink":
		return "blob"
	case "submodule":
		return "commit"
	default:
		return t
	}
}

type gitCommit struct {
	Sha    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

func parseHistory(body []byte) ([]Commit, error) {
	var list []gitCommit
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parse git history failed: %w", err)
	}
	out := make([]Commit, 0, len(list))
	for _, c := range list {
		if c.Sha == "" {
			continue
		}
		out = append(out, Commit{Sha: c.Sha, Message: c.Commit.Message, Author: c.Commit.Author.Name, Date: c.Commit.Author.Date})
	}
	return out, nil
}

type gitBranch struct {
	Name   string `json:"name"`
	Commit struct {
		Sha string `json:"sha"`
	} `json:"commit"`
}

func parseBranches(body []byte) ([]Branch, error) {
	var list []gitBranch
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parse git branches failed: %w", err)
	}
	out := make([]Branch, 0, len(list))
	for _, b := range list {
		if b.Name == "" {
			continue
		}
		out = append(out, Branch{Name: b.Name})
	}
	return out, nil
}
