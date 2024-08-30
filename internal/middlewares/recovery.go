package middlewares

import (
	"context"
	"net/http"

	"github.com/PicPay/lib-go-instrumentation/interfaces"
	logger "github.com/PicPay/lib-go-logger/v2"
	"github.com/gin-gonic/gin"
)

func RecoverMiddleware(ctx context.Context, instrument interfaces.Instrument, log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Recovered from panic: %s", r)
				histogramOperations(ctx, "panic", "error", instrument)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal Server Error",
					"message": "An unexpected error occurred",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

func histogramOperations(ctx context.Context, action string, status string, instrument interfaces.Instrument) {
	metricLabels := map[string]any{
		"action": action,
		"status": status,
	}
	counter := instrument.StartInt64Counter(ctx, "chatpicpay_panic_recovery_histogram", "custom panic recovery metrics")
	counter.Add(ctx, 1, metricLabels)
}