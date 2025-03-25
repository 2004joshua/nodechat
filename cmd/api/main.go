package main

import (
	"os"

	api "github.com/2004joshua/nodechat/pkg/api"
	"github.com/2004joshua/nodechat/pkg/pubsub"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	client := pubsub.NewClient(redisAddr)
	api.NewRouter(client).Run(":8081")
}
