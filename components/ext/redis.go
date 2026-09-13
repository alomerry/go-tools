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
		client, err := redis.NewRedisClient(os.Getenv(cons.REDIS_DSN))
		if err != nil {
			return err
		}
		r.redis = client
		return nil
	}

	client, err := redis.NewRedisClient(Apollo().GetRedisConfig().Uri)
	if err != nil {
		return err
	}
	r.redis = client
	return nil
}

// client 返回底层 redis 客户端。两种情况显式返回错误而非裸解引用触发 nil
// pointer panic：
//   - 接收者 nil：扩展未加载（NewRedisExtension 未执行，ExtRedis 未注册到扩展链）；
//   - r.redis 为 nil：扩展已加载但 Init 未调用。
func (r *RedisExtension) client() (*redis2.Client, error) {
  if r == nil {
    return nil, errors.New("redis extension not loaded: ExtRedis is not registered (check extension config)")
  }
  if r.redis == nil {
    return nil, errors.New("redis extension not initialized: RedisExtension.Init() has not been called (check ExtRedis config)")
  }
  return r.redis, nil
}

func (r *RedisExtension) Set(ctx context.Context, key string, value any) error {
  cli, err := r.client()
  if err != nil {
    return err
  }
  return cli.Set(ctx, key, value, 0).Err()
}

func (r *RedisExtension) SetEx(ctx context.Context, key string, value any, expiration time.Duration) error {
  cli, err := r.client()
  if err != nil {
    return err
  }
  return cli.SetEx(ctx, key, value, expiration).Err()
}

func (r *RedisExtension) HSet(ctx context.Context, key string, value ...any) error {
  cli, err := r.client()
  if err != nil {
    return err
  }
  return cli.HSet(ctx, key, value...).Err()
}

func (r *RedisExtension) Get(ctx context.Context, key string) (Getter, error) {
  cli, err := r.client()
  if err != nil {
    return nil, err
  }
  val, err := cli.Get(ctx, key).Result()
  if err != nil {
    return nil, err
  }
  return getter{val}, nil
}

// HSetField sets a single hash field. This is the field-scoped counterpart of
// the variadic HSet above, for callers that store one field at a time under a
// hash key (e.g. per-bucket aggregation state).
func (r *RedisExtension) HSetField(ctx context.Context, key, field string, value any) error {
  cli, err := r.client()
  if err != nil {
    return err
  }
  return cli.HSet(ctx, key, field, value).Err()
}

// HGet returns the value of a single hash field. It returns ("", nil) when the
// field does not exist (redis.Nil is treated as a normal miss, not an error).
func (r *RedisExtension) HGet(ctx context.Context, key, field string) (string, error) {
  cli, err := r.client()
  if err != nil {
    return "", err
  }
  val, err := cli.HGet(ctx, key, field).Result()
  if errors.Is(err, redis2.Nil) {
    return "", nil
  }
  return val, err
}

// HGetAll returns all fields and values of a hash key as a map. An empty map is
// returned when the key does not exist.
func (r *RedisExtension) HGetAll(ctx context.Context, key string) (map[string]string, error) {
  cli, err := r.client()
  if err != nil {
    return nil, err
  }
  return cli.HGetAll(ctx, key).Result()
}

// Del removes one or more keys. Returns the number of keys removed.
func (r *RedisExtension) Del(ctx context.Context, keys ...string) (int64, error) {
  cli, err := r.client()
  if err != nil {
    return 0, err
  }
  return cli.Del(ctx, keys...).Result()
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
