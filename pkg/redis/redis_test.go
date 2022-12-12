package redis_test

import (
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v9"
	wrappedRedis "github.com/maingockien01/proxy/pkg/redis"
	"github.com/stretchr/testify/require"
)

var redisServer *miniredis.Miniredis

var redisClient *wrappedRedis.RedisClient

func setup() {
	redisServer = mockRedis()
	client := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	redisClient = &wrappedRedis.RedisClient{
		Client: client,
	}
}

func mockRedis() *miniredis.Miniredis {
	s, err := miniredis.Run()

	if err != nil {
		panic(err)
	}

	return s
}

func teardown() {
	redisServer.Close()
}

func TestGetNil(t *testing.T) {
	setup()
	defer teardown()

	_, err := redisClient.Get("no_key")

	require.True(t, wrappedRedis.IsKeyNotExist(err))
}

func TestGetSuccess(t *testing.T) {
	setup()
	defer teardown()

	err := redisServer.Set("key", "value")

	if err != nil {
		panic(err)
	}

	val, getErr := redisClient.Get("key")

	require.Nil(t, getErr)

	require.Equal(t, val, "value")
}

func TestSetSuccess(t *testing.T) {
	setup()
	defer teardown()

	err := redisClient.Set("key", "value")

	require.Nil(t, err)

	redisServer.CheckGet(t, "key", "value")
}
