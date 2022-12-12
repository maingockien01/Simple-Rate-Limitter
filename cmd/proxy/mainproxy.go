package main

import (
	"fmt"
	"os"

	"github.com/maingockien01/proxy/pkg/server"
)

func main() {
	redisUrl := os.Getenv("REDIS_URL")
	fmt.Println("Start proxy...")
	proxy := server.NewServer(redisUrl)

	proxy.Serve("http://backend:80")
}
