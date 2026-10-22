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
	"strings"

	"github.com/alomerry/go-tools/components/http"
	"github.com/alomerry/go-tools/components/http/opts/req"
)

// Hook webhook 配置条目（github/gitee 归一后同形）：
//   - Name：github 为上游 name（repo hooks 恒为 "web"），gitee 无该字段固定
//     "webhook"，仅展示用途，调用方不得做语义判断。
//   - URL：webhook 目标地址（github config.url；gitee url）。
//   - Active：是否生效。gitee v5 无停用概念恒 true；github 透传 active。
//   - Events：触发事件名（github 透传 events 数组；gitee 由布尔位派生
//     push/tag_push/issues/note 中为 true 的项，事件名对齐 gitee 字段命名）。
//   - CreatedAt：创建时间原文字符串（两家时区/格式不同，不解析只透传）。
type Hook struct {
	ID        int64
	Name      string
	URL       string
	Active    bool
	Events    []string
	CreatedAt string
}

// Delivery 单次 webhook 调用记录（github only）。Status 为上游原文（"OK" /
// "failed to deliver" 等），StatusCode 为上游响应码；DurationSeconds 为上游
// 耗时（秒）。ID 为数字 delivery id（重发入参），GUID 为事件唯一标识。
type Delivery struct {
	ID              int64
	GUID            string
	Event           string
	Action          string
	Status          string
	StatusCode      int
	DeliveredAt     string
	Redelivered     bool
	DurationSeconds float64
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

// ListDeliveries 列 webhook 调用记录（github only，分页由模板透给上游）。
func ListDeliveries(ctx context.Context, q Query, tmpl string) ([]Delivery, error) {
	if q.Owner == "" || q.Repo == "" {
		return nil, fmt.Errorf("owner/repo are required")
	}
	if q.HookID <= 0 {
		return nil, fmt.Errorf("hook_id is required")
	}
	if !SupportsDeliveries(q.Provider) {
		return nil, fmt.Errorf("provider %s does not support webhook deliveries", q.Provider)
	}
	body, err := doFetch(ctx, q, tmpl)
	if err != nil {
		return nil, err
	}
	return parseDeliveries(body)
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
		return &UpstreamError{StatusCode: code, Snippet: snippet(body)}
	}
	return nil
}

// --- hooks 上游解析（github/gitee 形状不同，分派后各自一套 struct） ---

type githubHook struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
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
		out = append(out, Hook{ID: h.ID, Name: h.Name, URL: h.Config.URL, Active: h.Active, Events: h.Events, CreatedAt: h.CreatedAt})
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
// （Name 固定 "webhook"、Active 恒 true，见 Hook 注释），事件由布尔位派生。
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
		out = append(out, Hook{ID: h.ID, Name: "webhook", URL: h.URL, Active: true, Events: events, CreatedAt: h.CreatedAt})
	}
	return out, nil
}

// --- deliveries 上游解析（github only） ---

type githubDelivery struct {
	ID          int64   `json:"id"`
	GUID        string  `json:"guid"`
	Event       string  `json:"event"`
	Action      string  `json:"action"`
	Status      string  `json:"status"`
	StatusCode  int     `json:"status_code"`
	DeliveredAt string  `json:"delivered_at"`
	Redelivered bool    `json:"redelivered"`
	Duration    float64 `json:"duration"`
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
		out = append(out, Delivery{ID: d.ID, GUID: d.GUID, Event: d.Event, Action: d.Action, Status: d.Status, StatusCode: d.StatusCode, DeliveredAt: d.DeliveredAt, Redelivered: d.Redelivered, DurationSeconds: d.Duration})
	}
	return out, nil
}
