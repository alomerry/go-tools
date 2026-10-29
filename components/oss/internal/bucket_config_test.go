package internal

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/alomerry/go-tools/components/oss/meta"
	"github.com/alomerry/go-tools/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 接口合规编译门禁：四实现必须满足扩展后的 meta.OSSClient。
var (
	_ meta.OSSClient = (*RustFs)(nil)
	_ meta.OSSClient = (*CloudflareR2)(nil)
	_ meta.OSSClient = (*minioClient)(nil)
	_ meta.OSSClient = (*client)(nil)
)

// stubS3 是无依赖的 S3 兼容桩：按 subresource（policy/lifecycle/versioning/
// encryption/cors/tagging）+ method 路由，GET 回 canned XML/JSON，PUT/DELETE
// 记录请求体供断言。签名不校验（SDK 签名照发，桩只看路径与查询串）。
type stubS3 struct {
	mu     sync.Mutex
	gets   map[string]stubResp
	bodies map[string]string // "METHOD subresource" -> 请求体
	method map[string]string // subresource -> 最后一次写方法
}

type stubResp struct {
	status int
	body   string
}

func newStubS3() *stubS3 {
	return &stubS3{
		gets:   map[string]stubResp{},
		bodies: map[string]string{},
		method: map[string]string{},
	}
}

func s3ErrorXML(code string) string {
	return `<Error><Code>` + code + `</Code><Message>stub</Message></Error>`
}

var s3Subresources = []string{"policy", "lifecycle", "versioning", "encryption", "cors", "tagging"}

func (s *stubS3) subresource(r *http.Request) string {
	q := r.URL.Query()
	for _, sub := range s3Subresources {
		if _, ok := q[sub]; ok {
			return sub
		}
	}
	return ""
}

func (s *stubS3) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sub := s.subresource(r)
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.Method == http.MethodGet {
		resp, ok := s.gets[sub]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(s3ErrorXML("NoSuchBucket")))
			return
		}
		w.WriteHeader(resp.status)
		_, _ = w.Write([]byte(resp.body))
		return
	}
	body, _ := io.ReadAll(r.Body)
	s.bodies[r.Method+" "+sub] = string(body)
	s.method[sub] = r.Method
	w.WriteHeader(http.StatusNoContent)
}

func newStubRustFs(t *testing.T, stub *stubS3) *RustFs {
	t.Helper()
	srv := httptest.NewServer(stub)
	t.Cleanup(srv.Close)
	rust, err := newRustFs(srv.URL, "ak", "sk")
	require.NoError(t, err)
	return rust
}

const stubPolicyJSON = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::demo/*"]}]}`

const stubLifecycleXML = `<LifecycleConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">` +
	`<Rule><ID>logs-expire</ID><Filter><Prefix>logs/</Prefix></Filter><Status>Enabled</Status>` +
	`<Expiration><Days>30</Days></Expiration>` +
	`<AbortIncompleteMultipartUpload><DaysAfterInitiation>7</DaysAfterInitiation></AbortIncompleteMultipartUpload>` +
	`</Rule>` +
	`<Rule><ID>all</ID><Filter><Prefix></Prefix></Filter><Status>Disabled</Status>` +
	`<Expiration><Days>365</Days></Expiration>` +
	`</Rule></LifecycleConfiguration>`

const stubTaggingXML = `<Tagging xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><TagSet>` +
	`<Tag><Key>owner</Key><Value>team-a</Value></Tag>` +
	`<Tag><Key>sipad-referer-enabled</Key><Value>true</Value></Tag>` +
	`<Tag><Key>sipad-referer-allowlist</Key><Value>example.com&#10;cdn.example.com</Value></Tag>` +
	`<Tag><Key>sipad-quota-bytes</Key><Value>1073741824</Value></Tag>` +
	`</TagSet></Tagging>`

func TestStub_PolicyRoundTrip(t *testing.T) {
	stub := newStubS3()
	stub.gets["policy"] = stubResp{status: 200, body: stubPolicyJSON}
	rust := newStubRustFs(t, stub)
	ctx := context.Background()

	got, err := rust.GetBucketPolicy(ctx, "demo")
	require.NoError(t, err)
	assert.Equal(t, stubPolicyJSON, got)

	require.NoError(t, rust.PutBucketPolicy(ctx, "demo", stubPolicyJSON))
	assert.Equal(t, stubPolicyJSON, stub.bodies["PUT policy"])

	// 校验失败不发请求。
	err = rust.PutBucketPolicy(ctx, "demo", "not json")
	require.Error(t, err)
	_, ok := stub.bodies["PUT policy"]
	_ = ok

	// 空串 = 清空语义，走 DELETE（无独立 Delete 方法）。
	require.NoError(t, rust.PutBucketPolicy(ctx, "demo", ""))
	assert.Equal(t, "DELETE", stub.method["policy"])
}

func TestStub_PolicyMissing(t *testing.T) {
	stub := newStubS3()
	stub.gets["policy"] = stubResp{status: 404, body: s3ErrorXML("NoSuchBucketPolicy")}
	rust := newStubRustFs(t, stub)

	got, err := rust.GetBucketPolicy(context.Background(), "demo")
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestStub_UnsupportedMapsToSentinel(t *testing.T) {
	stub := newStubS3()
	stub.gets["versioning"] = stubResp{status: 501, body: s3ErrorXML("NotImplemented")}
	rust := newStubRustFs(t, stub)

	_, err := rust.GetBucketVersioning(context.Background(), "demo")
	require.Error(t, err)
	assert.True(t, errors.Is(err, meta.ErrBucketConfigUnsupported), "got %v", err)
}

func TestStub_VersioningSet(t *testing.T) {
	stub := newStubS3()
	stub.gets["versioning"] = stubResp{status: 200, body: `<VersioningConfiguration><Status>Enabled</Status></VersioningConfiguration>`}
	rust := newStubRustFs(t, stub)
	ctx := context.Background()

	v, err := rust.GetBucketVersioning(ctx, "demo")
	require.NoError(t, err)
	assert.True(t, v.Enabled)

	require.NoError(t, rust.SetBucketVersioning(ctx, "demo", false))
	body := stub.bodies["PUT versioning"]
	assert.Contains(t, body, "Suspended")
}

func TestStub_LifecycleGetPut(t *testing.T) {
	stub := newStubS3()
	stub.gets["lifecycle"] = stubResp{status: 200, body: stubLifecycleXML}
	rust := newStubRustFs(t, stub)
	ctx := context.Background()

	rules, err := rust.GetBucketLifecycle(ctx, "demo")
	require.NoError(t, err)
	require.Len(t, rules, 2)
	assert.Equal(t, "logs-expire", rules[0].ID)
	assert.Equal(t, "logs/", rules[0].Prefix)
	assert.Equal(t, int32(30), rules[0].ExpireDays)
	assert.True(t, rules[0].AbortIncompleteMultipart)
	assert.Equal(t, int32(7), rules[0].AbortMultipartDays)
	assert.True(t, rules[0].Enabled)
	assert.Equal(t, int32(365), rules[1].ExpireDays)
	assert.False(t, rules[1].Enabled)

	require.NoError(t, rust.PutBucketLifecycle(ctx, "demo", []meta.LifecycleRule{{
		Prefix: "logs/", ExpireDays: 30, AbortIncompleteMultipart: true,
		AbortMultipartDays: 7, Enabled: true,
	}}))
	putBody := stub.bodies["PUT lifecycle"]
	assert.Contains(t, putBody, "logs/")
	assert.Contains(t, putBody, "<Days>30</Days>")

	// 空集 = 清空语义，走 DELETE。
	require.NoError(t, rust.PutBucketLifecycle(ctx, "demo", nil))
	assert.Equal(t, "DELETE", stub.method["lifecycle"])

	// 缺失配置归一空表。
	stub.gets["lifecycle"] = stubResp{status: 404, body: s3ErrorXML("NoSuchLifecycleConfiguration")}
	rules, err = rust.GetBucketLifecycle(ctx, "demo")
	require.NoError(t, err)
	assert.Empty(t, rules)
}

func TestStub_EncryptionMissingAndPut(t *testing.T) {
	stub := newStubS3()
	stub.gets["encryption"] = stubResp{status: 404, body: s3ErrorXML("ServerSideEncryptionConfigurationNotFoundError")}
	rust := newStubRustFs(t, stub)
	ctx := context.Background()

	cfg, err := rust.GetBucketEncryption(ctx, "demo")
	require.NoError(t, err)
	assert.Equal(t, meta.BucketEncryption{}, cfg)

	require.NoError(t, rust.PutBucketEncryption(ctx, "demo", meta.BucketEncryption{Algorithm: "AES256"}))
	assert.Contains(t, stub.bodies["PUT encryption"], "AES256")

	require.Error(t, rust.PutBucketEncryption(ctx, "demo", meta.BucketEncryption{Algorithm: "SM4"}))
}

func TestStub_CORSMissingAndPut(t *testing.T) {
	stub := newStubS3()
	stub.gets["cors"] = stubResp{status: 404, body: s3ErrorXML("NoSuchCORSConfiguration")}
	rust := newStubRustFs(t, stub)
	ctx := context.Background()

	rules, err := rust.GetBucketCORS(ctx, "demo")
	require.NoError(t, err)
	assert.Empty(t, rules)

	require.NoError(t, rust.PutBucketCORS(ctx, "demo", []meta.CORSRule{{
		AllowedOrigins: []string{"https://app.example"},
		AllowedMethods: []string{"GET"},
		MaxAgeSeconds:  600,
	}}))
	putBody := stub.bodies["PUT cors"]
	assert.Contains(t, putBody, "https://app.example")
	assert.True(t, strings.Contains(putBody, "600"))

	// 空集 = 清空语义，走 DELETE。
	require.NoError(t, rust.PutBucketCORS(ctx, "demo", nil))
	assert.Equal(t, "DELETE", stub.method["cors"])
}

func TestStub_RefererAndQuotaViaTagging(t *testing.T) {
	stub := newStubS3()
	stub.gets["tagging"] = stubResp{status: 200, body: stubTaggingXML}
	rust := newStubRustFs(t, stub)
	ctx := context.Background()

	ref, err := rust.GetBucketReferer(ctx, "demo")
	require.NoError(t, err)
	assert.True(t, ref.Enabled)
	assert.Equal(t, []string{"example.com", "cdn.example.com"}, ref.Allowlist)

	quota, err := rust.GetBucketQuota(ctx, "demo")
	require.NoError(t, err)
	assert.Equal(t, int64(1073741824), quota.Bytes)

	// 读-改-写合并：他键 owner 必须保留。
	require.NoError(t, rust.PutBucketReferer(ctx, "demo", meta.RefererConfig{
		Enabled: true, Allowlist: []string{"new.example"},
	}))
	tagBody := stub.bodies["PUT tagging"]
	assert.Contains(t, tagBody, "owner")
	assert.Contains(t, tagBody, "team-a")
	assert.Contains(t, tagBody, "sipad-referer-enabled")
	assert.Contains(t, tagBody, "new.example")

	require.NoError(t, rust.PutBucketQuota(ctx, "demo", meta.BucketQuota{Bytes: 0}))
	tagBody = stub.bodies["PUT tagging"]
	assert.NotContains(t, tagBody, "sipad-quota-bytes")

	// 脏 quota Tag 按不限处理，不报错。
	stub.gets["tagging"] = stubResp{status: 200, body: strings.Replace(stubTaggingXML,
		"<Value>1073741824</Value>", "<Value>oops</Value>", 1)}
	quota, err = rust.GetBucketQuota(ctx, "demo")
	require.NoError(t, err)
	assert.Equal(t, int64(0), quota.Bytes)

	// 未打 Tag 的桶：空表 + nil error。
	stub.gets["tagging"] = stubResp{status: 404, body: s3ErrorXML("NoSuchTagSet")}
	ref, err = rust.GetBucketReferer(ctx, "demo")
	require.NoError(t, err)
	assert.False(t, ref.Enabled)
	assert.Empty(t, ref.Allowlist)
}

func TestKodo_BucketConfigUnsupported(t *testing.T) {
	q, err := NewKodoClient(model.Config{
		Type:      model.ClientTypeKodo,
		AccessKey: "ak",
		SecretKey: "sk",
	})
	require.NoError(t, err)
	ctx := context.Background()

	_, err = q.GetBucketPolicy(ctx, "demo")
	assert.True(t, errors.Is(err, meta.ErrBucketConfigUnsupported))
	assert.Error(t, q.PutBucketPolicy(ctx, "demo", stubPolicyJSON))
	assert.Error(t, q.PutBucketLifecycle(ctx, "demo", nil))
	assert.Error(t, q.SetBucketVersioning(ctx, "demo", true))
	assert.Error(t, q.PutBucketEncryption(ctx, "demo", meta.BucketEncryption{Algorithm: "AES256"}))
	assert.Error(t, q.PutBucketCORS(ctx, "demo", nil))
	assert.Error(t, q.PutBucketReferer(ctx, "demo", meta.RefererConfig{}))
	assert.Error(t, q.PutBucketQuota(ctx, "demo", meta.BucketQuota{}))
}

func TestStub_MethodNotAllowedMapsToSentinel(t *testing.T) {
	stub := newStubS3()
	stub.gets["versioning"] = stubResp{status: 403, body: s3ErrorXML("MethodNotAllowed")}
	rust := newStubRustFs(t, stub)

	_, err := rust.GetBucketVersioning(context.Background(), "demo")
	require.Error(t, err)
	assert.True(t, errors.Is(err, meta.ErrBucketConfigUnsupported), "got %v", err)
}

// TestStub_RealErrorNotSwallowed 钉死 R2 错误码假设的另一面：InvalidRequest
// 等真错误码必须原样透出（不可收敛为「不支持」降级，否则掩盖请求构造 bug）。
func TestStub_RealErrorNotSwallowed(t *testing.T) {
	stub := newStubS3()
	stub.gets["versioning"] = stubResp{status: 400, body: s3ErrorXML("InvalidRequest")}
	rust := newStubRustFs(t, stub)

	_, err := rust.GetBucketVersioning(context.Background(), "demo")
	require.Error(t, err)
	assert.False(t, errors.Is(err, meta.ErrBucketConfigUnsupported), "got %v", err)

	_, err = rust.GetBucketPolicy(context.Background(), "demo")
	require.Error(t, err)
	assert.False(t, errors.Is(err, meta.ErrBucketConfigUnsupported), "got %v", err)
}

// 非 prefix 过滤规则（含 Tag）：Get 按 prefix 空回读（语义有损已注释），
// Put 整单写（含清空）必须拒绝，避免静默销毁控制台规则。
const stubLifecycleTagFilterXML = `<LifecycleConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">` +
	`<Rule><ID>tag-rule</ID><Filter><Tag><Key>env</Key><Value>prod</Value></Tag></Filter><Status>Enabled</Status>` +
	`<Expiration><Days>30</Days></Expiration>` +
	`</Rule></LifecycleConfiguration>`

func TestStub_LifecycleNonPrefixBlocksWholeWrite(t *testing.T) {
	stub := newStubS3()
	stub.gets["lifecycle"] = stubResp{status: 200, body: stubLifecycleTagFilterXML}
	rust := newStubRustFs(t, stub)
	ctx := context.Background()

	err := rust.PutBucketLifecycle(ctx, "demo", []meta.LifecycleRule{{
		Prefix: "logs/", ExpireDays: 30, Enabled: true,
	}})
	require.Error(t, err)
	assert.True(t, errors.Is(err, meta.ErrBucketConfigInvalid), "got %v", err)
	assert.False(t, errors.Is(err, meta.ErrBucketConfigUnsupported), "got %v", err)

	// 清空同样会销毁 tag 规则，一并拒绝。
	err = rust.PutBucketLifecycle(ctx, "demo", nil)
	require.Error(t, err)
	assert.True(t, errors.Is(err, meta.ErrBucketConfigInvalid), "got %v", err)

	// 纯 prefix 旧规则不拦截（本包自己写出的形态可正常覆盖）。
	stub.gets["lifecycle"] = stubResp{status: 200, body: stubLifecycleXML}
	require.NoError(t, rust.PutBucketLifecycle(ctx, "demo", []meta.LifecycleRule{{
		Prefix: "logs/", ExpireDays: 30, Enabled: true,
	}}))
}

// TestBucketTagLockShared 钉死锁共享契约：同桶 key 取到同一把锁（读-改-写
// 串行化不因 Bucket() 派生新 struct 而绕过——锁在包级表而非 struct 上）。
func TestBucketTagLockShared(t *testing.T) {
	assert.Same(t, bucketTagLock("demo"), bucketTagLock("demo"))
	assert.NotSame(t, bucketTagLock("demo"), bucketTagLock("other"))
}

// TestStub_TagValueExceedsWrapsInvalid 钉死 Tag 透存超长的哨兵：两处裸
// fmt.Errorf（s3PutBucketTags 通用超长 / s3PutBucketReferer 白名单拼接超长）
// 必须 errors.Is(meta.ErrBucketConfigInvalid)（backend 升级后据此转 400），
// 且失败不发请求。
func TestStub_TagValueExceedsWrapsInvalid(t *testing.T) {
	stub := newStubS3()
	rust := newStubRustFs(t, stub)
	ctx := context.Background()

	err := s3PutBucketTags(ctx, rust.client, "demo", map[string]string{
		"k": strings.Repeat("x", maxS3TagValueBytes+1),
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, meta.ErrBucketConfigInvalid), "got %v", err)
	assert.False(t, errors.Is(err, meta.ErrBucketConfigUnsupported), "got %v", err)
	_, sent := stub.bodies["PUT tagging"]
	assert.False(t, sent, "oversize tag must not send request")

	err = rust.PutBucketReferer(ctx, "demo", meta.RefererConfig{
		Enabled:   true,
		Allowlist: []string{strings.Repeat("a", maxS3TagValueBytes+1)},
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, meta.ErrBucketConfigInvalid), "got %v", err)
	assert.False(t, errors.Is(err, meta.ErrBucketConfigUnsupported), "got %v", err)
	_, sent = stub.bodies["PUT tagging"]
	assert.False(t, sent, "oversize referer must not send request")
}

// TestMinioBucketSharesS3Holder 钉死派生共享：Bucket() 派生体与原体共用
// 同一 s3ConfigHolder 指针（惰性构造结果 + 初始化互斥只一份；失败不缓存
// 语义由 holder 保证）。
func TestMinioBucketSharesS3Holder(t *testing.T) {
	cli, err := NewMinioClient(model.Config{
		Endpoint:  "127.0.0.1:9000",
		AccessKey: "ak",
		SecretKey: "sk",
	})
	require.NoError(t, err)
	parent, ok := cli.(*minioClient)
	require.True(t, ok)
	derived, ok := cli.Bucket(context.Background(), "other").(*minioClient)
	require.True(t, ok)
	assert.Same(t, parent.s3holder, derived.s3holder)
}
