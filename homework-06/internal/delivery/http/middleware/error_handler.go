package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-06/internal/entity"
	"go.uber.org/zap"
)

func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		ginErr := c.Errors.Last()
		err := ginErr.Err
		requestID, _ := c.Get(RequestIDKey)

		status := http.StatusInternalServerError
		message := http.StatusText(status)

		switch {
		case ginErr.Type == gin.ErrorTypeBind:
			status = http.StatusBadRequest
			message = err.Error()
		case errors.Is(err, entity.ErrAlreadyExists):
			status = http.StatusConflict
			message = err.Error()
		case errors.Is(err, entity.ErrNotFound):
			status = http.StatusNotFound
			message = err.Error()
		case errors.Is(err, entity.ErrUnauthorized):
			status = http.StatusUnauthorized
			message = err.Error()
		case errors.Is(err, entity.ErrServiceUnavailable):
			status = http.StatusServiceUnavailable
			message = err.Error()
		}

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
