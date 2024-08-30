package middlewares

import (
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/gin-gonic/gin"
)

func GetUserIdFromHeader(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/panic" {
			c.Next()
			return
		}

		userId := c.GetHeader("user_id")

		if userId == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Request.Header.Set("user_id", userId)

		c.Next()
	}
}
