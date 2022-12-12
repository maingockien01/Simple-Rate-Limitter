package ratelimiter

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/maingockien01/proxy/pkg/redis"
)

type TokenBucket struct {
	key                  string
	rate                 int64
	maxTokens            int64
	lastRefillTimestamps time.Time
	currentTokens        int64
	lockMutex            sync.Mutex
	apiPath              string
}

func NewTokenBucket(rate int64, maxTokens int64, key string, apiPath string) *TokenBucket {
	return &TokenBucket{
		key:                  key,
		rate:                 rate,
		maxTokens:            maxTokens,
		lastRefillTimestamps: time.Now(),
		currentTokens:        maxTokens,
		apiPath:              apiPath,
	}
}

func (tb *TokenBucket) FetchLocked(client *redis.RedisClient) {
	lock := tb.Lock(client)
	defer lock.Release(context.Background())
	tb.Fetch(client)
}

func (tb *TokenBucket) Fetch(client *redis.RedisClient) {
	var unixMicroLastRefill int64
	lastRefillString, err := client.Get(fmt.Sprintf("%s_lastRefillTimestamps", tb.key))

	if redis.IsKeyNotExist(err) {
		unixMicroLastRefill = time.Now().UnixMicro()
		client.Set(fmt.Sprintf("%s_lastRefillTimestamps", tb.key), strconv.FormatInt(unixMicroLastRefill, 10))
	} else {
		val, formatErr := strconv.Atoi(lastRefillString)

		if formatErr != nil {
			panic(formatErr)
		}

		unixMicroLastRefill = int64(val)
	}

	tb.lastRefillTimestamps = time.UnixMicro(unixMicroLastRefill)

	redisCurrentTokens, tokenErr := client.Get(fmt.Sprintf("%s_currentTokens", tb.key))

	if redis.IsKeyNotExist(tokenErr) {
		client.Set(fmt.Sprintf("%s_currentTokens", tb.key), strconv.FormatInt(tb.currentTokens, 10))
	} else {
		val, formatErr := strconv.Atoi(redisCurrentTokens)

		if formatErr != nil {
			panic(formatErr)
		}

		tb.currentTokens = int64(val)
	}
}

func (tb *TokenBucket) Lock(client *redis.RedisClient) *redislock.Lock {
	lock, err := client.Locker.Obtain(context.Background(), tb.key, 1*time.Second, &redislock.Options{
		RetryStrategy: redislock.LinearBackoff(100 * time.Millisecond),
	})

	if err == redislock.ErrNotObtained {
		panic(err)
	}

	return lock
}

func (tb *TokenBucket) Push(client *redis.RedisClient) error {
	err := client.Set(fmt.Sprintf("%s_lastRefillTimestamps", tb.key), strconv.FormatInt(tb.lastRefillTimestamps.UnixMicro(), 10))

	if err != nil {
		return err
	}

	err = client.Set(fmt.Sprintf("%s_currentTokens", tb.key), strconv.FormatInt(tb.currentTokens, 10))

	if err != nil {
		return err
	}

	return nil
}

func (tb *TokenBucket) refill() {
	now := time.Now()

	end := time.Since(tb.lastRefillTimestamps)

	tokensTobeAdded := (end.Nanoseconds() * int64(tb.rate)) / 1000000000
	tb.currentTokens = int64(math.Min(float64(tb.currentTokens+tokensTobeAdded), float64(tb.maxTokens)))
	tb.lastRefillTimestamps = now
}

func (tb *TokenBucket) IsRequestAllowed(tokens int64) bool {
	tb.lockMutex.Lock()
	defer tb.lockMutex.Unlock()
	tb.refill()

	if tb.currentTokens >= tokens {
		tb.currentTokens = tb.currentTokens - tokens

		return true
	}

	return false
}

func (tb *TokenBucket) IsRequestAllowedRedis(tokens int64, client *redis.RedisClient) bool {
	lock := tb.Lock(client)
	defer lock.Release(context.Background())
	// tb.lockMutex.Lock()
	// defer tb.lockMutex.Unlock()

	tb.Fetch(client)
	tb.refill()

	if tb.currentTokens >= tokens {
		tb.currentTokens = tb.currentTokens - tokens

		tb.Push(client)
		return true
	}

	return false
}
