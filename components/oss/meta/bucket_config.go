package meta

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// 桶配置（BucketPolicy/Lifecycle/Versioning/Encryption/CORS/Referer/Quota）
// 的共享类型与纯校验。校验函数均为纯函数（无网络 I/O），供各实现与单测复用；
// 服务端（homelab-backend）与前端（sipad）另有各自口径的校验，本包只收敛
// S3 语义侧的不变量。

// ErrBucketConfigUnsupported 表示后端不支持该配置项（如 R2 的版本控制、
// 七牛 kodo 全系）：调用方据此渲染「降级/不支持」徽标，而非按失败处理。
var ErrBucketConfigUnsupported = errors.New("bucket config is not supported by this backend")

// ErrBucketConfigInvalid 表示桶配置请求体未通过服务端校验（JSON 语法、
// 体积、规则合法性等）：调用方（homelab-backend ossBucketSetErr）据此映射
// 为 InvalidArgument（400），而非 Internal（500）。所有 Validate* 函数的
// 错误均以 %w 包裹本哨兵返回，调用方用 errors.Is 判定。
var ErrBucketConfigInvalid = errors.New("bucket config is invalid")

// MaxBucketPolicyBytes 是桶策略 JSON 的服务端上限（S3 PutBucketPolicy 上限
// 20KB；此处 64KB 为入库前防误传，真实约束由存储侧执行）。
const MaxBucketPolicyBytes = 64 * 1024

// MaxLifecycleRules 是单桶生命周期规则数上限（S3 上限 1000；常用个位数，
// 此处 100 为防误传）。
const MaxLifecycleRules = 100

// MaxCORSRules 是单桶 CORS 规则数上限（S3 上限 100）。
const MaxCORSRules = 100

// MaxRefererAllowlist 是防盗链白名单条目上限（应用层语义，Tag 透存有长度
// 约束，见 internal 实现注释）。
const MaxRefererAllowlist = 100

// LifecycleRule 是生命周期规则的 S3 子集（需求 Q3：prefix + 过期天数 +
// 清理分片开关；S3 的 filter/tag/transition/glide 语义本期不暴露）。
type LifecycleRule struct {
	// ID 规则标识（空则服务端自动生成 "rule-<index>"，便于幂等复写）。
	ID string
	// Prefix 生效前缀（空 = 整桶）。
	Prefix string
	// ExpireDays 对象过期天数（>0，S3 Expiration.Days）。
	ExpireDays int32
	// AbortIncompleteMultipart 是否清理未完成的分片上传。
	AbortIncompleteMultipart bool
	// AbortMultipartDays 分片上传发起后多少天清理（Abort=true 时 >0）。
	AbortMultipartDays int32
	// Enabled 规则开关（S3 Status Enabled/Disabled）。
	Enabled bool
}

// BucketVersioning 是版本控制状态（S3 只有 Enabled/Suspended 两态；
// Suspended 语义 = Enabled=false，即「关」后历史版本仍保留可读）。
type BucketVersioning struct {
	Enabled bool
}

// BucketEncryption 是服务端加密配置（S3 SSE-S3/SSE-KMS 子集；rustfs 支持
// SSE-S3，R2 不支持 → ErrBucketConfigUnsupported 降级）。
type BucketEncryption struct {
	// Algorithm 加密算法："AES256"（SSE-S3）|"aws:kms"（SSE-KMS）；空 = 未配置。
	Algorithm string
	// KMSKeyID 仅 Algorithm=aws:kms 时有效。
	KMSKeyID string
}

// CORSRule 是单条 CORS 规则（S3 原生；R2 支持）。
type CORSRule struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	ExposeHeaders  []string
	// MaxAgeSeconds 预检缓存秒数（>=0）。
	MaxAgeSeconds int32
}

// RefererConfig 是防盗链配置。S3 无原生防盗链 API，本包定义应用层语义：
// 以 Bucket Tagging 透存（键见 TagReferer* 常量），由网关/前端在 presign
// 与公开读链路执行 Referer 检查。Tagging 不可用的后端返回
// ErrBucketConfigUnsupported。
type RefererConfig struct {
	Enabled bool
	// Allowlist Referer 白名单（域名或前缀，空 = 拒绝全部外链）。
	Allowlist []string
}

// TagReferer* / TagQuotaBytes 是 Referer/Quota 在 Bucket Tagging 中的透存键。
// PutBucketTagging 为全量覆盖语义，实现侧必须读-改-写合并，不得丢弃他键。
const (
	TagRefererEnabled = "sipad-referer-enabled"
	TagRefererAllow   = "sipad-referer-allowlist"
	TagQuotaBytes     = "sipad-quota-bytes"
	refererAllowSep   = "\n"
)

// BucketQuota 是桶配额。S3 无原生桶配额 API，本包定义应用层软限制语义：
// 以 Bucket Tagging 透存字节数，仅用于展示与上传前检查，不做服务端硬拦截
// （S3 presigned PUT 无法约束实际写入字节，见 homelab-backend 注释）；
// R2 同为软限制（note 由调用方展示）。
type BucketQuota struct {
	// Bytes 配额字节数（0 = 不限）。
	Bytes int64
}

// ValidatePolicyJSON 校验桶策略 JSON（语法 + 体积）。Principal "*" 的风险
// 提示由前端在保存前二次确认承担，后端/网关不解析语义。失败错误均包裹
// ErrBucketConfigInvalid（调用方 errors.Is 转 400）。
func ValidatePolicyJSON(policy string) error {
	if len(policy) > MaxBucketPolicyBytes {
		return fmt.Errorf("%w: policy exceeds %d bytes", ErrBucketConfigInvalid, MaxBucketPolicyBytes)
	}
	if strings.TrimSpace(policy) == "" {
		return fmt.Errorf("%w: policy is empty", ErrBucketConfigInvalid)
	}
	var v any
	if err := json.Unmarshal([]byte(policy), &v); err != nil {
		return fmt.Errorf("%w: policy is not valid JSON: %v", ErrBucketConfigInvalid, err)
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("%w: policy must be a JSON object", ErrBucketConfigInvalid)
	}
	if _, ok := obj["Statement"]; !ok {
		return fmt.Errorf("%w: policy must contain Statement", ErrBucketConfigInvalid)
	}
	return nil
}

// ValidateLifecycleRules 校验生命周期规则集（空集合法 = 清空语义，由调用
// 方映射为 DeleteBucketLifecycle）。失败错误均包裹 ErrBucketConfigInvalid。
func ValidateLifecycleRules(rules []LifecycleRule) error {
	if len(rules) > MaxLifecycleRules {
		return fmt.Errorf("%w: too many lifecycle rules: %d", ErrBucketConfigInvalid, len(rules))
	}
	seen := make(map[string]struct{}, len(rules))
	for i, r := range rules {
		if r.ExpireDays <= 0 {
			return fmt.Errorf("%w: rule[%d]: expireDays must be > 0", ErrBucketConfigInvalid, i)
		}
		if r.ExpireDays > 36500 {
			return fmt.Errorf("%w: rule[%d]: expireDays exceeds 100 years", ErrBucketConfigInvalid, i)
		}
		if r.AbortIncompleteMultipart && r.AbortMultipartDays <= 0 {
			return fmt.Errorf("%w: rule[%d]: abortMultipartDays must be > 0 when abort is enabled", ErrBucketConfigInvalid, i)
		}
		if r.ID != "" {
			if _, dup := seen[r.ID]; dup {
				return fmt.Errorf("%w: rule[%d]: duplicate id %q", ErrBucketConfigInvalid, i, r.ID)
			}
			seen[r.ID] = struct{}{}
		}
	}
	return nil
}

// ValidateEncryption 校验服务端加密配置。失败错误均包裹
// ErrBucketConfigInvalid。
func ValidateEncryption(cfg BucketEncryption) error {
	switch cfg.Algorithm {
	case "", "AES256":
		if cfg.KMSKeyID != "" {
			return fmt.Errorf("%w: kmsKeyId requires algorithm aws:kms", ErrBucketConfigInvalid)
		}
	case "aws:kms":
	default:
		return fmt.Errorf("%w: unsupported algorithm %q", ErrBucketConfigInvalid, cfg.Algorithm)
	}
	return nil
}

// ValidateCORSRules 校验 CORS 规则集（空集合法 = 清空语义）。失败错误均
// 包裹 ErrBucketConfigInvalid。
func ValidateCORSRules(rules []CORSRule) error {
	if len(rules) > MaxCORSRules {
		return fmt.Errorf("%w: too many cors rules: %d", ErrBucketConfigInvalid, len(rules))
	}
	for i, r := range rules {
		if len(r.AllowedOrigins) == 0 {
			return fmt.Errorf("%w: cors rule[%d]: allowedOrigins is empty", ErrBucketConfigInvalid, i)
		}
		if len(r.AllowedMethods) == 0 {
			return fmt.Errorf("%w: cors rule[%d]: allowedMethods is empty", ErrBucketConfigInvalid, i)
		}
		if r.MaxAgeSeconds < 0 {
			return fmt.Errorf("%w: cors rule[%d]: maxAgeSeconds must be >= 0", ErrBucketConfigInvalid, i)
		}
	}
	return nil
}

// ValidateReferer 校验防盗链配置（白名单条目去空；换行拒绝——条目经
// JoinRefererAllowlist 以换行分隔透存，含换行会破坏编解码；Tag value 长度
// 约束在 internal 透存侧执行，此处只做条数、空值与换行收敛）。
// 失败错误均包裹 ErrBucketConfigInvalid。
func ValidateReferer(cfg RefererConfig) error {
	if len(cfg.Allowlist) > MaxRefererAllowlist {
		return fmt.Errorf("%w: too many referer allowlist entries: %d", ErrBucketConfigInvalid, len(cfg.Allowlist))
	}
	for i, allow := range cfg.Allowlist {
		if strings.TrimSpace(allow) == "" {
			return fmt.Errorf("%w: referer allowlist[%d] is empty", ErrBucketConfigInvalid, i)
		}
		if strings.Contains(allow, "\n") {
			return fmt.Errorf("%w: referer allowlist[%d] must not contain newline", ErrBucketConfigInvalid, i)
		}
	}
	return nil
}

// JoinRefererAllowlist/SplitRefererAllowlist 是白名单在 Tag value 中的
// 编解码（换行分隔；条目本身不得含换行，由 ValidateReferer 拒绝）。
func JoinRefererAllowlist(allow []string) string {
	return strings.Join(allow, refererAllowSep)
}

// SplitRefererAllowlist 解码 Tag value（空串 = 空表；条目去空去重保序）。
func SplitRefererAllowlist(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, part := range strings.Split(s, refererAllowSep) {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

// ValidateQuota 校验桶配额（负数拒绝；0 = 不限）。失败错误包裹
// ErrBucketConfigInvalid。
func ValidateQuota(q BucketQuota) error {
	if q.Bytes < 0 {
		return fmt.Errorf("%w: quota bytes must be >= 0", ErrBucketConfigInvalid)
	}
	return nil
}
