package ext

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alomerry/go-tools/components/redis"
	"github.com/alomerry/go-tools/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	redis2 "github.com/redis/go-redis/v9"
)

// RedisExtSuite 对 RedisExt 的每个方法进行测试，连接真实的 Redis 实例。遵循项目的
// 集成测试约定（参见 components/redis/redis_test.go）：手动构造 ext 实例并注入真实的
// *redis.Client，绕过 Init()，因此不依赖 apollo。每个测试用 t.Name() 派生的前缀
// 隔离 key，并在结束时清理。当本机没有 Redis 时，使用 `go test -short` 可跳过。
func TestRedisExtSuite(t *testing.T) {
	suite.Run(t, new(RedisExtSuite))
}

type RedisExtSuite struct {
	test.BaseSuite
	ext *RedisExtension
}

// redisTestURL 允许通过 REDIS_URL 环境变量覆盖 redis 地址。env 包没有单机版
// GetRedisDSN，因此回退到 components/redis/redis_test.go 使用的空字符串。空 URL 会导致
// redis.ParseURL 失败，所以这种情况下 SetupSuite 会跳过该 suite。
func redisTestURL() string {
	if u := os.Getenv("REDIS_URL"); u != "" {
		return u
	}
	return ""
}

func (s *RedisExtSuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("skipping redis integration test in short mode")
	}

	url := redisTestURL()
	if url == "" {
		s.T().Skip("REDIS_URL not set, skipping redis integration test")
	}

	// 手动构造 ext 实例并注入真实客户端。有意不调用 Init()：它会从 apollo 拉取配置。
	client, err := redis.NewRedisClient(url)
	if err != nil {
		s.T().Fatalf("invalid redis url: %v", err)
	}
	s.ext = &RedisExtension{
		redis: client,
	}
}

// key 构造每个测试唯一的 key，使测试之间互不冲突，也不受残留数据影响。
func (s *RedisExtSuite) key(name string) string {
	return "exttest:" + name
}

func (s *RedisExtSuite) cleanup(ctx context.Context, keys ...string) {
	if _, err := s.ext.Del(ctx, keys...); err != nil {
		s.T().Logf("cleanup Del failed: %v", err)
	}
}

func (s *RedisExtSuite) TestSet() {
	ctx := context.Background()
	k := s.key("TestSet")
	defer s.cleanup(ctx, k)

	assert.NoError(s.T(), s.ext.Set(ctx, k, "hello"))

	g, err := s.ext.Get(ctx, k)
	assert.NoError(s.T(), err)
	assert.False(s.T(), g.Empty())
	assert.Equal(s.T(), "hello", g.String())
}

func (s *RedisExtSuite) TestSetEx() {
	ctx := context.Background()
	k := s.key("TestSetEx")
	defer s.cleanup(ctx, k)

	assert.NoError(s.T(), s.ext.SetEx(ctx, k, "ttl-value", time.Second))

	g, err := s.ext.Get(ctx, k)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "ttl-value", g.String())

	// 等待过期并确认 key 已消失。
	time.Sleep(1200 * time.Millisecond)
	g2, err := s.ext.Get(ctx, k)
	assert.NoError(s.T(), err)
	assert.True(s.T(), g2.Empty())
}

func (s *RedisExtSuite) TestHSet() {
	ctx := context.Background()
	k := s.key("TestHSet")
	defer s.cleanup(ctx, k)

	assert.NoError(s.T(), s.ext.HSet(ctx, k, "f1", "v1", "f2", "v2"))

	all, err := s.ext.HGetAll(ctx, k)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), map[string]string{"f1": "v1", "f2": "v2"}, all)
}

func (s *RedisExtSuite) TestHSetField() {
	ctx := context.Background()
	k := s.key("TestHSetField")
	defer s.cleanup(ctx, k)

	assert.NoError(s.T(), s.ext.HSetField(ctx, k, "field", "value"))

	v, err := s.ext.HGet(ctx, k, "field")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "value", v)
}

func (s *RedisExtSuite) TestGet() {
	ctx := context.Background()
	k := s.key("TestGet")
	defer s.cleanup(ctx, k)

	assert.NoError(s.T(), s.ext.Set(ctx, k, "get-value"))

	g, err := s.ext.Get(ctx, k)
	assert.NoError(s.T(), err)
	assert.False(s.T(), g.Empty())
	assert.Equal(s.T(), "get-value", g.String())

	// 缺失的 key -> Empty 的 getter，不报错。
	missing, err := s.ext.Get(ctx, k+":nope")
	assert.NoError(s.T(), err)
	assert.True(s.T(), missing.Empty())
}

func (s *RedisExtSuite) TestHGet() {
	ctx := context.Background()
	k := s.key("TestHGet")
	defer s.cleanup(ctx, k)

	assert.NoError(s.T(), s.ext.HSetField(ctx, k, "f", "hv"))

	v, err := s.ext.HGet(ctx, k, "f")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "hv", v)

	// 缺失的字段被视为正常的未命中：返回 ("", nil)，而非报错。
	v2, err := s.ext.HGet(ctx, k, "missing")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "", v2)
}

func (s *RedisExtSuite) TestHGetAll() {
	ctx := context.Background()
	k := s.key("TestHGetAll")
	defer s.cleanup(ctx, k)

	assert.NoError(s.T(), s.ext.HSet(ctx, k, "a", "1", "b", "2"))

	all, err := s.ext.HGetAll(ctx, k)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), map[string]string{"a": "1", "b": "2"}, all)

	// 缺失的 key -> 空 map，不报错。
	empty, err := s.ext.HGetAll(ctx, k+":nope")
	assert.NoError(s.T(), err)
	assert.Empty(s.T(), empty)
}

func (s *RedisExtSuite) TestDel() {
	ctx := context.Background()
	k := s.key("TestDel")
	assert.NoError(s.T(), s.ext.Set(ctx, k, "del-me"))

	n, err := s.ext.Del(ctx, k)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), n)

	// key 现已消失。
	g, err := s.ext.Get(ctx, k)
	assert.NoError(s.T(), err)
	assert.True(s.T(), g.Empty())

	// 删除不存在的 key 会移除 0 个 key。
	n2, err := s.ext.Del(ctx, k)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), n2)
}

// --- getter / Getter 纯逻辑单元测试（无需 redis） ---

// TestNotInitialized 验证扩展未初始化（redis == nil）时各方法返回错误而非
// 裸解引用触发 panic。
func TestNotInitialized(t *testing.T) {
	ctx := context.Background()
	ext := &RedisExtension{}

	_, err := ext.Get(ctx, "k")
	assert.Error(t, err)

	_, err = ext.HGet(ctx, "k", "f")
	assert.Error(t, err)

	_, err = ext.HGetAll(ctx, "k")
	assert.Error(t, err)

	_, err = ext.Del(ctx, "k")
	assert.Error(t, err)

	assert.Error(t, ext.Set(ctx, "k", "v"))
	assert.Error(t, ext.SetEx(ctx, "k", "v", time.Second))
	assert.Error(t, ext.HSet(ctx, "k", "f", "v"))
	assert.Error(t, ext.HSetField(ctx, "k", "f", "v"))

	// nil 接收者（扩展完全未加载）也必须返回 error 而非 panic。
	var nilExt *RedisExtension
	_, err = nilExt.Get(ctx, "k")
	assert.Error(t, err)
	assert.Error(t, nilExt.Set(ctx, "k", "v"))
}

func TestGetterEmpty(t *testing.T) {
	assert.True(t, getter{value: nil}.Empty())
	assert.False(t, getter{value: "x"}.Empty())
	assert.False(t, getter{value: 0}.Empty())
}

func TestGetterString(t *testing.T) {
	// 非 nil 的值通过 cast 转为字符串，与是否传入 default 参数无关。
	assert.Equal(t, "abc", getter{value: "abc"}.String())
	assert.Equal(t, "abc", getter{value: "abc"}.String("default"))

	// cast.ToString 会转换数值类型。
	assert.Equal(t, "123", getter{value: 123}.String())

	// nil 值且传入了 default 参数时返回 default。
	assert.Equal(t, "fallback", getter{value: nil}.String("fallback"))

	// nil 值且未传入 default 参数时返回 cast.ToString(nil)，即 ""。
	assert.Equal(t, "", getter{value: nil}.String())
}

// 编译期引用，确保 suite 被 -short 跳过时 redis2 导入不会被丢弃。
var _ redis2.Cmdable = (*redis2.Client)(nil)
