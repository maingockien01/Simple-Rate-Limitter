package redis

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v9"
)

var DEFAULT_EXPIRED_DURATION = 0 * time.Second

type RedisInterface interface {
	Set(key, value string) error
	Get(key string) (string, error)
}

func IsKeyNotExist(err error) bool {
	return err == redis.Nil
}

type RedisClient struct {
	Client *redis.Client
	Locker *redislock.Client
}

func New(addr, passwrod string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwrod,
		DB:       db,
	})

	locker := redislock.New(client)

	return &RedisClient{
		Client: client,
		Locker: locker,
	}, nil
}

func (rc *RedisClient) Set(key, value string) error {
	ctx := context.Background()
	err := rc.Client.Set(ctx, key, value, DEFAULT_EXPIRED_DURATION).Err()

	return err
}

func (rc *RedisClient) Get(key string) (string, error) {
	ctx := context.Background()
	val, err := rc.Client.Get(ctx, key).Result()

	return val, err
}
