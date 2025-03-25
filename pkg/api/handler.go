package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/2004joshua/nodechat/pkg/models"
	"github.com/2004joshua/nodechat/pkg/pubsub"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SendHandler(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var msg models.Message
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		payload, _ := json.Marshal(msg)
		if err := pubsub.Publish(rdb, msg.Recipient, payload); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func SubscribeHandler(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		ps := pubsub.Subscribe(rdb, userID)
		defer ps.Close()

		ch := ps.Channel()
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-ch; ok {
				c.SSEvent("message", msg.Payload)
				return true
			}
			return false
		})
	}
}
