package ext

import (
  "context"
  "errors"
  "os"
  "sync"
  "time"
  
  "github.com/alomerry/go-tools/components/redis"
  "github.com/alomerry/go-tools/static/cons"
  "github.com/alomerry/go-tools/static/env"
  redis2 "github.com/redis/go-redis/v9"
  "github.com/spf13/cast"
)

var (
  RedisExt  *RedisExtension
  cacheOnce sync.Once
)

func init() {
  Register(cons.ExtRedis, NewRedisExtension)
}

type RedisExtension struct {
  redis *redis2.Client
}

func NewRedisExtension() Ext {
  cacheOnce.Do(func() {
    RedisExt = &RedisExtension{}
  })
  return RedisExt
}

func (r *RedisExtension) Init(ctx context.Context) error {
  if env.Local() {
    r.redis = redis.NewRedisClient(os.Getenv(cons.REDIS_DSN))
    return nil
  }
  
  r.redis = redis.NewRedisClient(Apollo().GetRedisConfig().Uri)
  return nil
}

func (r *RedisExtension) Set(ctx context.Context, key string, value any) error {
  return r.redis.Set(ctx, key, value, 0).Err()
}

func (r *RedisExtension) SetEx(ctx context.Context, key string, value any, expiration time.Duration) error {
  return r.redis.SetEx(ctx, key, value, expiration).Err()
}

func (r *RedisExtension) HSet(ctx context.Context, key string, value ...any) error {
  return r.redis.HSet(ctx, key, value...).Err()
}

func (r *RedisExtension) Get(ctx context.Context, key string) (Getter, error) {
  val, err := r.redis.Get(ctx, key).Result()
  if err != nil {
    return nil, err
  }
  return getter{val}, nil
}

// HSetField sets a single hash field. This is the field-scoped counterpart of
// the variadic HSet above, for callers that store one field at a time under a
// hash key (e.g. per-bucket aggregation state).
func (r *RedisExtension) HSetField(ctx context.Context, key, field string, value any) error {
  return r.redis.HSet(ctx, key, field, value).Err()
}

// HGet returns the value of a single hash field. It returns ("", nil) when the
// field does not exist (redis.Nil is treated as a normal miss, not an error).
func (r *RedisExtension) HGet(ctx context.Context, key, field string) (string, error) {
  val, err := r.redis.HGet(ctx, key, field).Result()
  if errors.Is(err, redis2.Nil) {
    return "", nil
  }
  return val, err
}

// HGetAll returns all fields and values of a hash key as a map. An empty map is
// returned when the key does not exist.
func (r *RedisExtension) HGetAll(ctx context.Context, key string) (map[string]string, error) {
  return r.redis.HGetAll(ctx, key).Result()
}

// Del removes one or more keys. Returns the number of keys removed.
func (r *RedisExtension) Del(ctx context.Context, keys ...string) (int64, error) {
  return r.redis.Del(ctx, keys...).Result()
}

type Getter interface {
  String(args ...string) string
  Empty() bool
}

type getter struct {
  value any
}

func (g getter) Empty() bool {
  return g.value == nil
}

func (g getter) String(args ...string) string {
  if len(args) >= 1 && g.value == nil {
    return args[0]
  }
  
  return cast.ToString(g.value)
}
