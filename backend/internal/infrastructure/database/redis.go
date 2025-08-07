package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func NewRedisClient(config RedisConfig) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{Client: rdb}, nil
}

func (r *RedisClient) GetWithTenant(ctx context.Context, tenantID, key string) (string, error) {
	tenantKey := fmt.Sprintf("tenant:%s:%s", tenantID, key)
	return r.Get(ctx, tenantKey).Result()
}

func (r *RedisClient) SetWithTenant(ctx context.Context, tenantID, key string, value interface{}, expiration time.Duration) error {
	tenantKey := fmt.Sprintf("tenant:%s:%s", tenantID, key)
	return r.Set(ctx, tenantKey, value, expiration).Err()
}

func (r *RedisClient) DelWithTenant(ctx context.Context, tenantID, key string) error {
	tenantKey := fmt.Sprintf("tenant:%s:%s", tenantID, key)
	return r.Del(ctx, tenantKey).Err()
}

func (r *RedisClient) ExistsWithTenant(ctx context.Context, tenantID, key string) (bool, error) {
	tenantKey := fmt.Sprintf("tenant:%s:%s", tenantID, key)
	result, err := r.Exists(ctx, tenantKey).Result()
	return result > 0, err
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}
