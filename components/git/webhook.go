// webhook 能力（service-git-webhook）：ListHooks / ListDeliveries /
// RedeliverDelivery。取数与校验复用 git.go 的 doFetch/UpstreamError/2MB
// 上限口径；与浏览四能力的差异点：
//   - hooks 列表 github/gitee 返回形状不同（github: {id,name,config.url,
//     active,events}；gitee: {id,url,created_at,push_events,...} 布尔事件位），
//     无法像浏览四能力共用一套 parse，按 provider 内部分派解析，产出同形 Hook。
//   - deliveries/redeliver 仅 github 提供（gitee v5 无对应 API），除「空模板
//     fail-closed」外再做 provider 白名单，防误配模板打错平台。
//   - redeliver 是 POST（上游 204 无 body），空 body 不算失败——不复用
//     doFetch 的空 body 报错分支，走独立 doPost。
package git

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/alomerry/go-tools/components/http"
	"github.com/alomerry/go-tools/components/http/opts/req"
)

// Hook webhook 配置条目（github/gitee 归一后同形）：
//   - URL：webhook 目标地址（github config.url；gitee url）。
//   - Active：是否生效。gitee v5 无停用概念恒 true；github 透传 active。
//   - Events：触发事件名（github 透传 events 数组；gitee 由布尔位派生
//     push/tag_push/issues/note 中为 true 的项，事件名对齐 gitee 字段命名）。
//   - CreatedAt：创建时间原文字符串（两家时区/格式不同，不解析只透传）。
type Hook struct {
	ID        int64
	URL       string
	Active    bool
	Events    []string
	CreatedAt string
}

// Delivery 单次 webhook 调用记录（github only）。Status 为上游原文（"OK" /
// "failed to deliver" 等），StatusCode 为上游响应码。ID 为数字 delivery id
// （重发入参）；上游 guid/duration 等展示性字段不透出（无消费场景）。
type Delivery struct {
	ID          int64
	Event       string
	Action      string
	Status      string
	StatusCode  int
	DeliveredAt string
	Redelivered bool
}

// PostFetch 单次 HTTP POST 取数（redeliver 专用，变量化以便单测桩替换）；
// 契约同 FetchFunc。
var PostFetch FetchFunc = func(ctx context.Context, url, token string) (int, []byte, error) {
	resp, err := http.GetClient().Post(ctx, url, req.WithAuthToken(token))
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode(), resp.Bytes(), nil
}

// SupportsDeliveries 该 provider 是否支持 deliveries 查询与单条重发
// （gitee v5 无对应 API，github 独有能力）。
func SupportsDeliveries(provider string) bool {
	return provider == ProviderGithub
}

// ListHooks 列 webhook 配置（分页由模板的 {page}/{page_size} 透给上游；
// 形状两家不同，按 provider 内部分派解析）。
func ListHooks(ctx context.Context, q Query, tmpl string) ([]Hook, error) {
	if q.Owner == "" || q.Repo == "" {
		return nil, fmt.Errorf("owner/repo are required")
	}
	body, err := doFetch(ctx, q, tmpl)
	if err != nil {
		return nil, err
	}
	switch q.Provider {
	case ProviderGithub:
		return parseGithubHooks(body)
	case ProviderGitee:
		return parseGiteeHooks(body)
	default:
		return nil, fmt.Errorf("unsupported git provider %s", q.Provider)
	}
}

// ListDeliveries 列 webhook 调用记录（github only，cursor 分页）：上游
// List deliveries 端点仅支持 per_page/cursor（page 为非法键，传即 422），
// 故不复用通用 BuildURL——经 BuildDeliveriesURL 装配（模板残留 {page} 即
// fail-fast），返回体外带下一页游标（无下一页即空串；取自响应 Link 头
// rel="next"，上游单页/末页不下发 Link 头）。
func ListDeliveries(ctx context.Context, q Query, tmpl string) ([]Delivery, string, error) {
	if q.Owner == "" || q.Repo == "" {
		return nil, "", fmt.Errorf("owner/repo are required")
	}
	if q.HookID <= 0 {
		return nil, "", fmt.Errorf("hook_id is required")
	}
	if !SupportsDeliveries(q.Provider) {
		return nil, "", fmt.Errorf("provider %s does not support webhook deliveries", q.Provider)
	}
	target, err := BuildDeliveriesURL(tmpl, q)
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(q.Host) == "" {
		return nil, "", fmt.Errorf("git host is not configured for provider %s", q.Provider)
	}
	code, header, body, err := FetchHeaders(ctx, target, q.Token)
	if err != nil {
		return nil, "", err
	}
	if err := checkBrowseBody(code, body); err != nil {
		return nil, "", err
	}
	deliveries, err := parseDeliveries(body)
	if err != nil {
		return nil, "", err
	}
	return deliveries, parseNextCursor(header.Get("Link")), nil
}

// parseNextCursor 从 github 分页 Link 头提取 rel="next" 的 cursor 参数：
// `<url?per_page=20&cursor=abc>; rel="next", <url...>; rel="prev"`。
// 无 Link 头/无 next 段/cursor 为空（单页或末页）一律回空串，不报错。
func parseNextCursor(link string) string {
	for _, part := range strings.Split(link, ",") {
		segments := strings.Split(strings.TrimSpace(part), ";")
		if len(segments) < 2 {
			continue
		}
		next := false
		for _, s := range segments[1:] {
			if strings.TrimSpace(s) == `rel="next"` {
				next = true
				break
			}
		}
		if !next {
			continue
		}
		u, err := url.Parse(strings.Trim(strings.TrimSpace(segments[0]), "<>"))
		if err != nil {
			continue
		}
		if cursor := u.Query().Get("cursor"); cursor != "" {
			return cursor
		}
	}
	return ""
}

// RedeliverDelivery 重发单条 webhook 调用（github only，POST；上游 2xx 即
// 触发成功，204 无 body 属正常，不做空 body 报错）。
func RedeliverDelivery(ctx context.Context, q Query, tmpl string) error {
	if q.Owner == "" || q.Repo == "" {
		return fmt.Errorf("owner/repo are required")
	}
	if q.HookID <= 0 || q.DeliveryID <= 0 {
		return fmt.Errorf("hook_id/delivery_id are required")
	}
	if !SupportsDeliveries(q.Provider) {
		return fmt.Errorf("provider %s does not support webhook deliveries", q.Provider)
	}
	if tmpl == "" {
		return fmt.Errorf("provider %s has no path template for this operation", q.Provider)
	}
	if strings.TrimSpace(q.Host) == "" {
		return fmt.Errorf("git host is not configured for provider %s", q.Provider)
	}
	target, err := BuildURL(tmpl, q)
	if err != nil {
		return err
	}
	code, body, err := PostFetch(ctx, target, q.Token)
	if err != nil {
		return err
	}
	if code/100 != 2 {
		return newUpstreamError(code, body)
	}
	return nil
}

// --- hooks 上游解析（github/gitee 形状不同，分派后各自一套 struct） ---

type githubHook struct {
	ID     int64 `json:"id"`
	Active bool  `json:"active"`
	Config struct {
		URL string `json:"url"`
	} `json:"config"`
	Events    []string `json:"events"`
	CreatedAt string   `json:"created_at"`
}

func parseGithubHooks(body []byte) ([]Hook, error) {
	var list []githubHook
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parse git hooks failed: %w", err)
	}
	out := make([]Hook, 0, len(list))
	for _, h := range list {
		// ID 缺失条目跳过（限流错误 JSON 能过 unmarshal 但无 id，兜底不产出
		// 零值行；HTTP 状态校验是第一道防线，与 parseTree 同口径）。
		if h.ID == 0 {
			continue
		}
		out = append(out, Hook{ID: h.ID, URL: h.Config.URL, Active: h.Active, Events: h.Events, CreatedAt: h.CreatedAt})
	}
	return out, nil
}

type giteeHook struct {
	ID            int64  `json:"id"`
	URL           string `json:"url"`
	CreatedAt     string `json:"created_at"`
	PushEvents    bool   `json:"push_events"`
	TagPushEvents bool   `json:"tag_push_events"`
	IssuesEvents  bool   `json:"issues_events"`
	NoteEvents    bool   `json:"note_events"`
}

// parseGiteeHooks gitee v5 hooks 数组 → Hook：无 name/active 概念
// （Active 恒 true，见 Hook 注释），事件由布尔位派生。
func parseGiteeHooks(body []byte) ([]Hook, error) {
	var list []giteeHook
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parse git hooks failed: %w", err)
	}
	out := make([]Hook, 0, len(list))
	for _, h := range list {
		if h.ID == 0 {
			continue
		}
		events := make([]string, 0, 4)
		if h.PushEvents {
			events = append(events, "push")
		}
		if h.TagPushEvents {
			events = append(events, "tag_push")
		}
		if h.IssuesEvents {
			events = append(events, "issues")
		}
		if h.NoteEvents {
			events = append(events, "note")
		}
		out = append(out, Hook{ID: h.ID, URL: h.URL, Active: true, Events: events, CreatedAt: h.CreatedAt})
	}
	return out, nil
}

// --- deliveries 上游解析（github only） ---

type githubDelivery struct {
	ID          int64  `json:"id"`
	Event       string `json:"event"`
	Action      string `json:"action"`
	Status      string `json:"status"`
	StatusCode  int    `json:"status_code"`
	DeliveredAt string `json:"delivered_at"`
	Redelivered bool   `json:"redelivered"`
}

func parseDeliveries(body []byte) ([]Delivery, error) {
	var list []githubDelivery
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parse git deliveries failed: %w", err)
	}
	out := make([]Delivery, 0, len(list))
	for _, d := range list {
		if d.ID == 0 {
			continue
		}
		out = append(out, Delivery{ID: d.ID, Event: d.Event, Action: d.Action, Status: d.Status, StatusCode: d.StatusCode, DeliveredAt: d.DeliveredAt, Redelivered: d.Redelivered})
	}
	return out, nil
}
