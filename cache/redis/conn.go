package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	pool      *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = "123456"
)

// 创建redis连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   30,
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			// 连接redis
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			// 访问认证
			if _, err := c.Do("AUTH", redisPass); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := conn.Do("PING")
			return err
		},
	}
}

// 初始化一个连接池对象
func init() {
	pool = newRedisPool()
}

// 返回redis数据库连接对象
func RedisPool() *redis.Pool {
	return pool
}
