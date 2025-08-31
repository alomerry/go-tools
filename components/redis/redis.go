package redis

import (
	"github.com/redis/go-redis/v9"
	"log"
)

func NewRedisClient(url string) *redis.Client {
	// redis://<user>:<pass>@localhost:6379/<db>
	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Panicf("parse redis url failed, url: %v, err: %v", url, err.Error())
	}

	return redis.NewClient(opt)
}
