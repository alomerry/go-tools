package redis

import (
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(url string) *redis.Client {
	// redis://<user>:<pass>@localhost:6379/<db>
	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Panicf("parse redis url failed, url: %v, err: %v", url, err.Error())
	}

	return redis.NewClient(opt)
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
