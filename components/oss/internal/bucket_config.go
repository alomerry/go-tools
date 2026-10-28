package internal

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/alomerry/go-tools/components/oss/meta"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// 桶配置的 S3 共享实现（rust_fs.go 的 RustFs 与 r2.go 的 CloudflareR2 共用，
// 均持 *s3.Client）：BucketPolicy/Lifecycle/Versioning/Encryption/CORS 走
// S3 原生 API；Referer/Quota 为应用层语义，经 Bucket Tagging 透存
// （meta.TagReferer*/TagQuotaBytes，Put 全量覆盖故读-改-写合并）。
//
// 不存在语义归一为空值 + nil error（S3 以专有错误码表达「未配置」，调用方
// 不应将其视为失败）：
//   - policy 缺失 NoSuchBucketPolicy → ""；lifecycle 缺失
//     NoSuchLifecycleConfiguration → nil；encryption 缺失
//     ServerSideEncryptionConfigurationNotFoundError → 零值；cors 缺失
//     NoSuchCORSConfiguration → nil；tagging 缺失 NoSuchTagSet → 空表。
// 不支持语义归一 ErrBucketConfigUnsupported（R2 的版本控制/加密等返回
// NotImplemented/MethodNotAllowed，调用方降级展示）。

// asBucketConfigErr 将「后端不支持」类错误码归一为
// ErrBucketConfigUnsupported（调用方降级展示）；其余错误原样透出。
// 仅收敛 NotImplemented/MethodNotAllowed：InvalidRequest 为真错误（误吞会
// 掩盖请求构造 bug），NotSupported 非 S3 标准码，不收敛。
//
// R2 假设（未真机验证，单测钉死）：R2 文档与 S3 兼容层对不支持的桶配置
// API 回 NotImplemented（部分操作 MethodNotAllowed）；若真机回码不同，
// 按该码补 case 并同步更新 backend 探测注释。InvalidRequest 等真错误码
// 必须原样透出（见 TestStub_RealErrorNotSwallowed）。
func asBucketConfigErr(err error) error {
	if err == nil {
		return nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NotImplemented", "MethodNotAllowed":
			// R2 对不支持的桶配置 API 回 NotImplemented/MethodNotAllowed。
			return fmt.Errorf("%w: %s", meta.ErrBucketConfigUnsupported, apiErr.ErrorCode())
		}
	}
	return err
}

// isNotFoundCode 判定是否为「未配置」类错误码。
func isNotFoundCode(err error, codes ...string) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	for _, code := range codes {
		if apiErr.ErrorCode() == code {
			return true
		}
	}
	return false
}

// --- BucketPolicy ---

func s3GetBucketPolicy(ctx context.Context, client *s3.Client, bucket string) (string, error) {
	out, err := client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{Bucket: &bucket})
	if err != nil {
		if isNotFoundCode(err, "NoSuchBucketPolicy") {
			return "", nil
		}
		return "", asBucketConfigErr(err)
	}
	return aws.ToString(out.Policy), nil
}

func s3PutBucketPolicy(ctx context.Context, client *s3.Client, bucket, policyJSON string) error {
	if strings.TrimSpace(policyJSON) == "" {
		return s3DeleteBucketPolicy(ctx, client, bucket)
	}
	if err := meta.ValidatePolicyJSON(policyJSON); err != nil {
		return err
	}
	_, err := client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: &bucket,
		Policy: &policyJSON,
	})
	return asBucketConfigErr(err)
}

func s3DeleteBucketPolicy(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.DeleteBucketPolicy(ctx, &s3.DeleteBucketPolicyInput{Bucket: &bucket})
	return asBucketConfigErr(err)
}

// --- Lifecycle (Q3: prefix + 过期天数 + 清理分片开关) ---

func s3GetBucketLifecycle(ctx context.Context, client *s3.Client, bucket string) ([]meta.LifecycleRule, error) {
	out, err := client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{Bucket: &bucket})
	if err != nil {
		if isNotFoundCode(err, "NoSuchLifecycleConfiguration") {
			return nil, nil
		}
		return nil, asBucketConfigErr(err)
	}
	// 口径局限：本包只表达 prefixFilter 语义。控制台创建的非 prefix 过滤
	// 规则（tag/size/And）在此按 Prefix 空回读（语义有损），但 Put 侧整单
	// 写保护（s3LifecycleHasNonPrefixFilter）会拒绝覆盖，需去控制台编辑。
	rules := make([]meta.LifecycleRule, 0, len(out.Rules))
	for _, r := range out.Rules {
		mr := meta.LifecycleRule{
			ID:      aws.ToString(r.ID),
			Enabled: r.Status == types.ExpirationStatusEnabled,
		}
		if r.Filter != nil {
			mr.Prefix = aws.ToString(r.Filter.Prefix)
		} else {
			mr.Prefix = aws.ToString(r.Prefix)
		}
		if r.Expiration != nil {
			mr.ExpireDays = aws.ToInt32(r.Expiration.Days)
		}
		if r.AbortIncompleteMultipartUpload != nil {
			mr.AbortIncompleteMultipart = true
			mr.AbortMultipartDays = aws.ToInt32(r.AbortIncompleteMultipartUpload.DaysAfterInitiation)
		}
		rules = append(rules, mr)
	}
	return rules, nil
}

func s3PutBucketLifecycle(ctx context.Context, client *s3.Client, bucket string, rules []meta.LifecycleRule) error {
	if err := meta.ValidateLifecycleRules(rules); err != nil {
		return err
	}
	// 整单写保护：S3 Put 为全量覆盖。本包只表达 prefixFilter 语义；若桶上
	// 已有控制台创建的非 prefix 过滤规则（tag/size/And 组合），直接覆盖会
	// 静默销毁其语义，故拒绝整单写，请用户去控制台编辑。读失败（NoSuch*
	// 缺失/真错误）不拦截：缺失即无旧规则可破坏，真错误的 Put 同样会失败
	// 并原样透出。
	if blocked, reason := s3LifecycleHasNonPrefixFilter(ctx, client, bucket); blocked {
		return fmt.Errorf("%w: bucket has non-prefix lifecycle filter (%s), edit in console instead of whole-order overwrite", meta.ErrBucketConfigInvalid, reason)
	}
	if len(rules) == 0 {
		return s3DeleteBucketLifecycle(ctx, client, bucket)
	}
	s3rules := make([]types.LifecycleRule, 0, len(rules))
	for i, r := range rules {
		id := r.ID
		if id == "" {
			id = fmt.Sprintf("rule-%d", i+1)
		}
		sr := types.LifecycleRule{
			ID:         &id,
			Status:     types.ExpirationStatusDisabled,
			Filter:     &types.LifecycleRuleFilter{Prefix: &r.Prefix},
			Expiration: &types.LifecycleExpiration{Days: aws.Int32(r.ExpireDays)},
		}
		if r.Enabled {
			sr.Status = types.ExpirationStatusEnabled
		}
		if r.AbortIncompleteMultipart {
			days := r.AbortMultipartDays
			sr.AbortIncompleteMultipartUpload = &types.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: &days,
			}
		}
		s3rules = append(s3rules, sr)
	}
	_, err := client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket: &bucket,
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: s3rules,
		},
	})
	return asBucketConfigErr(err)
}

func s3DeleteBucketLifecycle(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.DeleteBucketLifecycle(ctx, &s3.DeleteBucketLifecycleInput{Bucket: &bucket})
	return asBucketConfigErr(err)
}

// s3LifecycleHasNonPrefixFilter 探测桶上是否已有非 prefix 过滤的生命周期
// 规则（控制台创建的 tag/size/And 组合过滤；S3 Filter 除 Prefix 外任一字段
// 非空即判定命中）。返回 (true, 原因) 时调用方拒绝整单覆盖写。
// 注意：仅做「正向命中」拦截；读失败一律回 (false, "") 不拦截。
func s3LifecycleHasNonPrefixFilter(ctx context.Context, client *s3.Client, bucket string) (bool, string) {
	out, err := client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{Bucket: &bucket})
	if err != nil {
		return false, ""
	}
	for i, r := range out.Rules {
		f := r.Filter
		if f == nil {
			// 无 Filter + 顶层 Prefix 为空以外的旧形态（S3 deprecated 字段）：
			// 本包 Get 侧按 Prefix 语义回读，不拦截。
			continue
		}
		if f.Tag != nil || f.And != nil || f.ObjectSizeGreaterThan != nil || f.ObjectSizeLessThan != nil {
			return true, fmt.Sprintf("rule[%d] uses tag/size filter", i)
		}
	}
	return false, ""
}

// --- Versioning ---

func s3GetBucketVersioning(ctx context.Context, client *s3.Client, bucket string) (meta.BucketVersioning, error) {
	out, err := client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: &bucket})
	if err != nil {
		return meta.BucketVersioning{}, asBucketConfigErr(err)
	}
	return meta.BucketVersioning{Enabled: out.Status == types.BucketVersioningStatusEnabled}, nil
}

func s3SetBucketVersioning(ctx context.Context, client *s3.Client, bucket string, enabled bool) error {
	status := types.BucketVersioningStatusSuspended
	if enabled {
		status = types.BucketVersioningStatusEnabled
	}
	_, err := client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: &bucket,
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: status,
		},
	})
	return asBucketConfigErr(err)
}

// --- Encryption ---

func s3GetBucketEncryption(ctx context.Context, client *s3.Client, bucket string) (meta.BucketEncryption, error) {
	out, err := client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: &bucket})
	if err != nil {
		if isNotFoundCode(err, "ServerSideEncryptionConfigurationNotFoundError") {
			return meta.BucketEncryption{}, nil
		}
		return meta.BucketEncryption{}, asBucketConfigErr(err)
	}
	cfg := meta.BucketEncryption{}
	if out.ServerSideEncryptionConfiguration != nil {
		for _, r := range out.ServerSideEncryptionConfiguration.Rules {
			if r.ApplyServerSideEncryptionByDefault == nil {
				continue
			}
			d := r.ApplyServerSideEncryptionByDefault
			cfg.Algorithm = string(d.SSEAlgorithm)
			cfg.KMSKeyID = aws.ToString(d.KMSMasterKeyID)
			break // 本期只暴露首条规则
		}
	}
	return cfg, nil
}

func s3PutBucketEncryption(ctx context.Context, client *s3.Client, bucket string, cfg meta.BucketEncryption) error {
	if err := meta.ValidateEncryption(cfg); err != nil {
		return err
	}
	if cfg.Algorithm == "" {
		return s3DeleteBucketEncryption(ctx, client, bucket)
	}
	def := &types.ServerSideEncryptionByDefault{
		SSEAlgorithm: types.ServerSideEncryption(cfg.Algorithm),
	}
	if cfg.Algorithm == "aws:kms" && cfg.KMSKeyID != "" {
		def.KMSMasterKeyID = &cfg.KMSKeyID
	}
	_, err := client.PutBucketEncryption(ctx, &s3.PutBucketEncryptionInput{
		Bucket: &bucket,
		ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
			Rules: []types.ServerSideEncryptionRule{
				{ApplyServerSideEncryptionByDefault: def},
			},
		},
	})
	return asBucketConfigErr(err)
}

func s3DeleteBucketEncryption(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.DeleteBucketEncryption(ctx, &s3.DeleteBucketEncryptionInput{Bucket: &bucket})
	return asBucketConfigErr(err)
}

// --- CORS ---

func s3GetBucketCORS(ctx context.Context, client *s3.Client, bucket string) ([]meta.CORSRule, error) {
	out, err := client.GetBucketCors(ctx, &s3.GetBucketCorsInput{Bucket: &bucket})
	if err != nil {
		if isNotFoundCode(err, "NoSuchCORSConfiguration") {
			return nil, nil
		}
		return nil, asBucketConfigErr(err)
	}
	rules := make([]meta.CORSRule, 0, len(out.CORSRules))
	for _, r := range out.CORSRules {
		rules = append(rules, meta.CORSRule{
			AllowedOrigins: r.AllowedOrigins,
			AllowedMethods: r.AllowedMethods,
			AllowedHeaders: r.AllowedHeaders,
			ExposeHeaders:  r.ExposeHeaders,
			MaxAgeSeconds:  aws.ToInt32(r.MaxAgeSeconds),
		})
	}
	return rules, nil
}

func s3PutBucketCORS(ctx context.Context, client *s3.Client, bucket string, rules []meta.CORSRule) error {
	if err := meta.ValidateCORSRules(rules); err != nil {
		return err
	}
	if len(rules) == 0 {
		return s3DeleteBucketCORS(ctx, client, bucket)
	}
	s3rules := make([]types.CORSRule, 0, len(rules))
	for _, r := range rules {
		s3rules = append(s3rules, types.CORSRule{
			AllowedOrigins: r.AllowedOrigins,
			AllowedMethods: r.AllowedMethods,
			AllowedHeaders: r.AllowedHeaders,
			ExposeHeaders:  r.ExposeHeaders,
			MaxAgeSeconds:  aws.Int32(r.MaxAgeSeconds),
		})
	}
	_, err := client.PutBucketCors(ctx, &s3.PutBucketCorsInput{
		Bucket: &bucket,
		CORSConfiguration: &types.CORSConfiguration{
			CORSRules: s3rules,
		},
	})
	return asBucketConfigErr(err)
}

func s3DeleteBucketCORS(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.DeleteBucketCors(ctx, &s3.DeleteBucketCorsInput{Bucket: &bucket})
	return asBucketConfigErr(err)
}

// --- Tagging 透存（Referer/Quota 应用层语义） ---
//
// S3 PutBucketTagging 为全量覆盖：写任一透存键前必须先读全量 TagSet 再合并
// 写回（读-改-写），否则会丢弃用户在控制台打的其他标签。Tag 键值约束
// （AWS 侧）：key ≤128 字符、value ≤256 字符；超长白名单/非法换行在写侧
// 拒绝并返回 descriptive error。
//
// 并发：读-改-写跨两次 S3 调用非原子，并发写同一桶会丢更新。包级按桶 key
// 加互斥（tagLocks）串行化同桶合并写——锁必须在共享包级而非各实现 struct
// 上：minioClient.Bucket()/RustFs.Bucket() 派生新 struct，若锁在 struct 上
// 会被派生绕过。跨端点同名桶会过度串行化（只降并发不降正确性，管理面低频
// 可接受）。不同桶互不阻塞。
const maxS3TagValueBytes = 256

// bucketTagLocks 是按桶 key 的互斥表（见上并发注释）。
var (
	bucketTagLocksMu sync.Mutex
	bucketTagLocks   = map[string]*sync.Mutex{}
)

// bucketTagLock 取（无则建）指定桶的合并写互斥锁。
func bucketTagLock(bucket string) *sync.Mutex {
	bucketTagLocksMu.Lock()
	defer bucketTagLocksMu.Unlock()
	if l, ok := bucketTagLocks[bucket]; ok {
		return l
	}
	l := &sync.Mutex{}
	bucketTagLocks[bucket] = l
	return l
}

func s3GetBucketTags(ctx context.Context, client *s3.Client, bucket string) (map[string]string, error) {
	out, err := client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{Bucket: &bucket})
	if err != nil {
		if isNotFoundCode(err, "NoSuchTagSet") {
			return map[string]string{}, nil
		}
		return nil, asBucketConfigErr(err)
	}
	tags := make(map[string]string, len(out.TagSet))
	for _, t := range out.TagSet {
		tags[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}
	return tags, nil
}

func s3PutBucketTags(ctx context.Context, client *s3.Client, bucket string, tags map[string]string) error {
	tagSet := make([]types.Tag, 0, len(tags))
	for k, v := range tags {
		if len(v) > maxS3TagValueBytes {
			return fmt.Errorf("%w: tag %q value exceeds %d bytes", meta.ErrBucketConfigInvalid, k, maxS3TagValueBytes)
		}
		k, v := k, v
		tagSet = append(tagSet, types.Tag{Key: &k, Value: &v})
	}
	_, err := client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
		Bucket:  &bucket,
		Tagging: &types.Tagging{TagSet: tagSet},
	})
	return asBucketConfigErr(err)
}

// s3MergeBucketTags 读-改-写：mutate 在全量副本上修改后写回（同桶串行化，
// 见包级 bucketTagLock 并发注释）。
func s3MergeBucketTags(ctx context.Context, client *s3.Client, bucket string, mutate func(map[string]string)) error {
	lock := bucketTagLock(bucket)
	lock.Lock()
	defer lock.Unlock()
	tags, err := s3GetBucketTags(ctx, client, bucket)
	if err != nil {
		return err
	}
	mutate(tags)
	return s3PutBucketTags(ctx, client, bucket, tags)
}

func s3GetBucketReferer(ctx context.Context, client *s3.Client, bucket string) (meta.RefererConfig, error) {
	tags, err := s3GetBucketTags(ctx, client, bucket)
	if err != nil {
		return meta.RefererConfig{}, err
	}
	return meta.RefererConfig{
		Enabled:   tags[meta.TagRefererEnabled] == "true",
		Allowlist: meta.SplitRefererAllowlist(tags[meta.TagRefererAllow]),
	}, nil
}

func s3PutBucketReferer(ctx context.Context, client *s3.Client, bucket string, cfg meta.RefererConfig) error {
	if err := meta.ValidateReferer(cfg); err != nil {
		return err
	}
	joined := meta.JoinRefererAllowlist(cfg.Allowlist)
	if len(joined) > maxS3TagValueBytes {
		return fmt.Errorf("%w: referer allowlist exceeds %d bytes in tag storage, reduce entries", meta.ErrBucketConfigInvalid, maxS3TagValueBytes)
	}
	enabled := "false"
	if cfg.Enabled {
		enabled = "true"
	}
	return s3MergeBucketTags(ctx, client, bucket, func(tags map[string]string) {
		tags[meta.TagRefererEnabled] = enabled
		if joined == "" {
			delete(tags, meta.TagRefererAllow)
		} else {
			tags[meta.TagRefererAllow] = joined
		}
	})
}

func s3GetBucketQuota(ctx context.Context, client *s3.Client, bucket string) (meta.BucketQuota, error) {
	tags, err := s3GetBucketTags(ctx, client, bucket)
	if err != nil {
		return meta.BucketQuota{}, err
	}
	raw := strings.TrimSpace(tags[meta.TagQuotaBytes])
	if raw == "" {
		return meta.BucketQuota{}, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 {
		// 脏 Tag（如手改控制台）不致命：按不限处理，写侧下次覆盖纠正。
		return meta.BucketQuota{}, nil
	}
	return meta.BucketQuota{Bytes: n}, nil
}

func s3PutBucketQuota(ctx context.Context, client *s3.Client, bucket string, quota meta.BucketQuota) error {
	if err := meta.ValidateQuota(quota); err != nil {
		return err
	}
	return s3MergeBucketTags(ctx, client, bucket, func(tags map[string]string) {
		if quota.Bytes == 0 {
			delete(tags, meta.TagQuotaBytes)
		} else {
			tags[meta.TagQuotaBytes] = strconv.FormatInt(quota.Bytes, 10)
		}
	})
}
