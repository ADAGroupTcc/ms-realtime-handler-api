package middlewares

import (
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/internal/clients/sessionClient"
	"github.com/gin-gonic/gin"
)

func Authenticate(sessionClient sessionClient.SessionClienter, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Token")
		requestId := c.GetHeader("X-Request-Id")

		if token == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userId, err := sessionClient.ValidateMobileToken(c, token, requestId, log)

		if err != nil {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Request.Header.Set("user_id", userId)

		c.Next()
	}
}
