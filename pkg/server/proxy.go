package server

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/maingockien01/proxy/pkg/ratelimiter"
	"github.com/maingockien01/proxy/pkg/redis"
	"github.com/maingockien01/proxy/pkg/ruler"
)

type ProxyServer struct {
	RedisClient  *redis.RedisClient
	RateLimiter  *ratelimiter.BucketRateLimiter
	LimiterRuler *ruler.Ruler
}

func NewServer(redisUrl string) *ProxyServer {
	redisClient, err := redis.New(redisUrl, "", 0)

	if err != nil {
		panic(err)
	}

	dir, _ := os.Getwd()
	path := filepath.Join(dir, "/configs/rules.json")

	limiterRuler := ruler.NewRuler(path, func(r *ruler.Ruler) {}, 2*time.Minute)

	limiterRuler.FetchFile()

	limiter := ratelimiter.NewRateLimiter(redisClient, limiterRuler)

	return &ProxyServer{
		RedisClient: redisClient,
		RateLimiter: limiter,
	}
}

func (s *ProxyServer) Serve(targetUrl string) {
	url, err := url.Parse(targetUrl)

	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	http.HandleFunc("/", s.Handler(proxy))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s *ProxyServer) Handler(reserveProxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		isAccept := s.RateLimiter.AcceptRequest(*req)
		if isAccept {
			reserveProxy.ServeHTTP(w, req)
		} else {
			//TODO: set up way to return data
			w.Header().Add("X-Ratelimit-Retry-After", "1s")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Please try again!"))
		}
	}
}
