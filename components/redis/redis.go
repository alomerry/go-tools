package redis

import (
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(url string) (*redis.Client, error) {
	// redis://<user>:<pass>@localhost:6379/<db>
	opt, err := redis.ParseURL(url)
	if err != nil {
		// 不打印原始 url（含明文密码），不 panic（构造失败交由调用方决策）
		return nil, fmt.Errorf("parse redis url failed: %w", err)
	}

	return redis.NewClient(opt), nil
}

var (
	redisKeyGenerator = &keyGenerator{}
)

func KeyGen() Generator {
	return redisKeyGenerator
}

type Generator interface {
	GenKey(category string, args ...string) string
}

type keyGenerator struct {
}

func (*keyGenerator) GenKey(category string, args ...string) string {
	return category + ":" + strings.Join(args, ":")
}
