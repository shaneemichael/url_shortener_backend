package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedis() *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return &RedisClient{Client: rdb}
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

func (r *RedisClient) SetKeyValue(ctx context.Context, key, value string) (bool, error) {
	ok, err := r.Client.SetNX(ctx, key, value, 0).Result()
	if err != nil {
		return false, fmt.Errorf("redis SETNX failed: %w", err)
	}
	return ok, nil
}

func (r *RedisClient) SetKeyValueWithTTL(ctx context.Context, key, value string, ttl int) (bool, error) {
	ok, err := r.Client.SetNX(ctx, key, value, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		return false, fmt.Errorf("redis SETNX failed: %w", err)
	}
	return ok, nil
}

func (r *RedisClient) GetKeyValue(ctx context.Context, key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}
