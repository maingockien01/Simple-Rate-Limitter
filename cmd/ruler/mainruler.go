package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/maingockien01/proxy/pkg/redis"
	"github.com/maingockien01/proxy/pkg/ruler"
)

func main() {
	redisUrl := os.Getenv("REDIS_URL")
	redisClient, err := redis.New(redisUrl, "", 0)

	if err != nil {
		panic(err)
	}
	dir, _ := os.Getwd()
	path := filepath.Join(dir, "/configs/rules.json")
	limiterRuler := ruler.NewRuler(path, func(r *ruler.Ruler) {}, 2*time.Minute)

	limiterRuler.FetchFile()
	limiterRuler.FetchFileAndPushRedisInterval(redisClient)

	for {
		select {}
	}
}
