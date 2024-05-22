package middlewares

import (
	"github.com/PicPay/ms-chatpicpay-websocket-handler-api/pkg/logs"
	"github.com/gin-gonic/gin"
)

func EnhanceLogger(jf *logs.JSONFormatter) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.GetHeader("X-Request-Id")
		jf.SetRequestID(requestId)
		c.Next()
	}
}
