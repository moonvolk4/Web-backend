package auth

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type RedisBlacklist struct{ rdb *redis.Client }

func NewRedisBlacklist(addr, password string) *RedisBlacklist {
	opt := &redis.Options{Addr: addr}
	if password != "" {
		opt.Password = password
	}
	return &RedisBlacklist{rdb: redis.NewClient(opt)}
}

func (rb *RedisBlacklist) IsBlacklisted(jti string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	val, err := rb.rdb.Get(ctx, rb.key(jti)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

func (rb *RedisBlacklist) Blacklist(jti string, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	d := time.Until(expiresAt)
	if d <= 0 {
		d = time.Hour // fallback small TTL
	}
	return rb.rdb.Set(ctx, rb.key(jti), "1", d).Err()
}

func (rb *RedisBlacklist) key(jti string) string { return "jwt:blacklist:" + jti }
