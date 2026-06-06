package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/entity"
	"go.uber.org/zap"
)

func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		requestID, _ := c.Get(RequestIDKey)

		status := http.StatusInternalServerError
		message := http.StatusText(status)

		logger.Error("request failed",
			zap.Error(err),
			zap.Int("status", status),
			zap.Any("request_id", requestID),
		)

		c.AbortWithStatusJSON(status, entity.ErrorResponse{
			Code:    int32(status),
			Message: message,
		})
	}
}
