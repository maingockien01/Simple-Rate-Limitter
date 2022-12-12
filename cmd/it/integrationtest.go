package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/maingockien01/proxy/pkg/redis"
)

func main() {
	itRedisConn()
	itProxyConn()
}

func itRedisConn() {
	redisClient, err := redis.New("redis:6379", "", 0)

	if err != nil {
		panic(err)
	}

	pingErr := redisClient.Client.Ping(context.Background()).Err()

	if pingErr != nil {
		panic(pingErr)
	}

	fmt.Println("Redis connection success")
}

func itProxyConn() {
	resp, err := http.Get("http://backend:80")

	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}
