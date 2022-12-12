package ratelimiter

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"

	"github.com/maingockien01/proxy/pkg/redis"
	"github.com/maingockien01/proxy/pkg/ruler"
)

type BucketRateLimiter struct {
	RedisClient   *redis.RedisClient
	ClientBuckets map[string]*TokenBucket
	Ruler         *ruler.Ruler
}

func NewRateLimiter(redisClient *redis.RedisClient, limitRuler *ruler.Ruler) *BucketRateLimiter {
	limiter := &BucketRateLimiter{
		RedisClient:   redisClient,
		ClientBuckets: make(map[string]*TokenBucket),
		Ruler:         limitRuler,
	}
	limitRuler.OnFetch = limiter.onRulerFetch
	limitRuler.FetchRedisInterval(redisClient)
	return limiter
}

func (limiter *BucketRateLimiter) onRulerFetch(limiterRuler *ruler.Ruler) {
	//Update tokens and rate

	for _, bucket := range limiter.ClientBuckets {
		rule := limiterRuler.GetRule(bucket.apiPath)
		bucket.rate = rule.Rate
		bucket.maxTokens = rule.MaxTokens
	}
}

func getKey(req http.Request) string {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		panic(err)
	}

	url := req.URL.Path
	keyString := fmt.Sprintf("%s-%s", ip, url)
	hash := sha256.Sum256([]byte(keyString))

	return fmt.Sprintf("%x", hash)
}

func (limiter *BucketRateLimiter) createNewBucket(req http.Request) *TokenBucket {
	key := getKey(req)
	reqPath := req.URL.Path

	if len(reqPath) == 0 {
		reqPath = "/"
	}

	rule := limiter.Ruler.GetRule(reqPath)

	return NewTokenBucket(rule.Rate, rule.MaxTokens, key, rule.ApiPath)

}

func (limiter *BucketRateLimiter) getBucket(req http.Request) *TokenBucket {
	key := getKey(req)

	if limiter.ClientBuckets[key] == nil {
		limiter.ClientBuckets[key] = limiter.createNewBucket(req)
	}

	return limiter.ClientBuckets[key]
}

func (limiter *BucketRateLimiter) AcceptRequest(req http.Request) bool {

	tokenBucket := limiter.getBucket(req)

	// return tokenBucket.IsRequestAllowed(1)
	return tokenBucket.IsRequestAllowedRedis(1, limiter.RedisClient)

}
