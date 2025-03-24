package main

import (
	"nodechat/pkg/api"
	"nodechat/pkg/pubsub"
	"os"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	client := pubsub.NewClient(redisAddr)
	api.NewRouter(client).Run(":8080")
}
