package middlewares

import (
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/gin-gonic/gin"
)

func GetUserIdFromHeader(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		headerString := ""
		for key, value := range c.Request.Header {
			headerString += key + ": " + value[0] + " | "
		}

		log.Info("Headers: " + headerString)

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
