package api

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func NewRouter(rdb *redis.Client) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/messages", SendHandler(rdb))
	r.GET("/subscribe/:userID", SubscribeHandler(rdb))
	return r
}
