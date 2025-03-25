package pubsub

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

func Publish(client *redis.Client, channel string, payload []byte) error {
	return client.Publish(ctx, channel, payload).Err()
}

func Subscribe(client *redis.Client, channel string) *redis.PubSub {
	return client.Subscribe(ctx, channel)
}
