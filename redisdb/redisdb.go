package redisdb

import (
	"github.com/gomodule/redigo/redis"
	"task/config"
	"time"
)

func NewRedisPool(cfg config.RedisConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.Host+":"+cfg.Port, redis.DialDatabase(cfg.DB))
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
