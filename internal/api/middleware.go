package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tareqpi/transfer-system/internal/logger"
	"go.uber.org/zap"
)

const headerRequestID = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(headerRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
			c.Request.Header.Set(headerRequestID, requestID)
		}
		c.Writer.Header().Set(headerRequestID, requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		requestID := c.GetString("request_id")

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.Int("status", status),
			zap.Duration("latency_ms", latency),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("bytes", size),
			zap.String("client_ip", c.ClientIP()),
		}

		if len(c.Errors) > 0 {
			logger.L().Error("request completed with errors", append(fields, zap.String("errors", c.Errors.String()))...)
		} else {
			logger.L().Info("request completed", fields...)
		}
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID := c.GetString("request_id")
				logger.L().Error("panic recovered", zap.Any("panic", rec), zap.String("request_id", requestID))
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
					RequestID: requestID,
					Error: ErrorObject{
						Code:    "internal_error",
						Message: http.StatusText(http.StatusInternalServerError),
					},
				})
			}
		}()
		c.Next()
	}
}
