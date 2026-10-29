package meta

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePolicyJSON(t *testing.T) {
	assert.Error(t, ValidatePolicyJSON(""))
	assert.Error(t, ValidatePolicyJSON("   "))
	assert.Error(t, ValidatePolicyJSON("not json"))
	assert.Error(t, ValidatePolicyJSON(`[1,2,3]`))
	assert.Error(t, ValidatePolicyJSON(`{"Version":"2012-10-17"}`))
	assert.NoError(t, ValidatePolicyJSON(`{"Version":"2012-10-17","Statement":[]}`))
	assert.Error(t, ValidatePolicyJSON(strings.Repeat("x", MaxBucketPolicyBytes+1)))
}

func TestValidateLifecycleRules(t *testing.T) {
	assert.NoError(t, ValidateLifecycleRules(nil)) // 清空语义
	assert.Error(t, ValidateLifecycleRules([]LifecycleRule{{Prefix: "a/", ExpireDays: 0}}))
	assert.Error(t, ValidateLifecycleRules([]LifecycleRule{{ExpireDays: 30, AbortIncompleteMultipart: true}}))
	assert.Error(t, ValidateLifecycleRules([]LifecycleRule{
		{ID: "dup", ExpireDays: 1},
		{ID: "dup", ExpireDays: 2},
	}))
	assert.Error(t, ValidateLifecycleRules([]LifecycleRule{{ExpireDays: 36501}}))
	assert.NoError(t, ValidateLifecycleRules([]LifecycleRule{{
		ID: "r1", Prefix: "logs/", ExpireDays: 30,
		AbortIncompleteMultipart: true, AbortMultipartDays: 7, Enabled: true,
	}}))
	many := make([]LifecycleRule, MaxLifecycleRules+1)
	for i := range many {
		many[i] = LifecycleRule{ExpireDays: 1}
	}
	assert.Error(t, ValidateLifecycleRules(many))
}

func TestValidateEncryption(t *testing.T) {
	assert.NoError(t, ValidateEncryption(BucketEncryption{}))
	assert.NoError(t, ValidateEncryption(BucketEncryption{Algorithm: "AES256"}))
	assert.NoError(t, ValidateEncryption(BucketEncryption{Algorithm: "aws:kms", KMSKeyID: "key-1"}))
	assert.NoError(t, ValidateEncryption(BucketEncryption{Algorithm: "aws:kms", KMSKeyID: ""}))
	assert.Error(t, ValidateEncryption(BucketEncryption{Algorithm: "AES256", KMSKeyID: "key-1"}))
	assert.Error(t, ValidateEncryption(BucketEncryption{Algorithm: "SM4"}))
}

func TestValidateCORSRules(t *testing.T) {
	assert.NoError(t, ValidateCORSRules(nil)) // 清空语义
	assert.Error(t, ValidateCORSRules([]CORSRule{{AllowedMethods: []string{"GET"}}}))
	assert.Error(t, ValidateCORSRules([]CORSRule{{
		AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET"}, MaxAgeSeconds: -1,
	}}))
	assert.NoError(t, ValidateCORSRules([]CORSRule{{
		AllowedOrigins: []string{"https://a.example"},
		AllowedMethods: []string{"GET", "PUT"},
		MaxAgeSeconds:  3600,
	}}))
}

func TestValidateReferer(t *testing.T) {
	assert.NoError(t, ValidateReferer(RefererConfig{}))
	assert.NoError(t, ValidateReferer(RefererConfig{Enabled: true, Allowlist: []string{"example.com"}}))
	assert.Error(t, ValidateReferer(RefererConfig{Allowlist: []string{"  "}}))
	assert.Error(t, ValidateReferer(RefererConfig{Allowlist: []string{"a\nb"}}))
	many := make([]string, MaxRefererAllowlist+1)
	for i := range many {
		many[i] = "a.example"
	}
	assert.Error(t, ValidateReferer(RefererConfig{Allowlist: many}))
}

func TestValidateQuota(t *testing.T) {
	assert.NoError(t, ValidateQuota(BucketQuota{}))
	assert.NoError(t, ValidateQuota(BucketQuota{Bytes: 1073741824}))
	assert.Error(t, ValidateQuota(BucketQuota{Bytes: -1}))
}

func TestRefererAllowlistRoundTrip(t *testing.T) {
	assert.Equal(t, "", JoinRefererAllowlist(nil))
	got := SplitRefererAllowlist("")
	assert.Empty(t, got)
	in := []string{"a.example", "b.example", "a.example", "  ", "c.example"}
	got = SplitRefererAllowlist(JoinRefererAllowlist(in))
	require.Equal(t, []string{"a.example", "b.example", "c.example"}, got)
}

// TestValidateErrorsWrapInvalid 钉死校验失败哨兵：所有 Validate* 的失败错误
// 必须 errors.Is(ErrBucketConfigInvalid)（backend 升级后据此转 400；
// 不支持哨兵保持独立，不可混淆）。
func TestValidateErrorsWrapInvalid(t *testing.T) {
	cases := []error{
		ValidatePolicyJSON("not json"),
		ValidatePolicyJSON(strings.Repeat("x", MaxBucketPolicyBytes+1)),
		ValidateLifecycleRules([]LifecycleRule{{ExpireDays: 0}}),
		ValidateLifecycleRules([]LifecycleRule{{ID: "d", ExpireDays: 1}, {ID: "d", ExpireDays: 2}}),
		ValidateEncryption(BucketEncryption{Algorithm: "SM4"}),
		ValidateEncryption(BucketEncryption{Algorithm: "", KMSKeyID: "k"}),
		ValidateCORSRules([]CORSRule{{AllowedOrigins: []string{"*"}, AllowedMethods: nil}}),
		ValidateReferer(RefererConfig{Allowlist: []string{"a\nb"}}),
		ValidateQuota(BucketQuota{Bytes: -1}),
	}
	for i, err := range cases {
		require.Error(t, err, "case %d", i)
		assert.True(t, errors.Is(err, ErrBucketConfigInvalid), "case %d: %v", i, err)
		assert.False(t, errors.Is(err, ErrBucketConfigUnsupported), "case %d", i)
	}
}
