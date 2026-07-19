package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	platformmiddleware "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/httpmiddleware"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/entity"
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
		status, message := http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
		switch {
		case ginErr.Type == gin.ErrorTypeBind, errors.Is(err, entity.ErrInvalidOperation):
			status, message = http.StatusBadRequest, err.Error()
		case errors.Is(err, entity.ErrNotFound):
			status, message = http.StatusNotFound, err.Error()
		case errors.Is(err, entity.ErrAlreadyExists):
			status, message = http.StatusConflict, err.Error()
		case errors.Is(err, entity.ErrInsufficientFunds):
			status, message = http.StatusConflict, err.Error()
		case errors.Is(err, entity.ErrUnauthorized):
			status, message = http.StatusUnauthorized, err.Error()
		case errors.Is(err, entity.ErrServiceUnavailable):
			status, message = http.StatusServiceUnavailable, err.Error()
		}
		requestID, _ := c.Get(platformmiddleware.RequestIDKey)
		logger.Error("request failed", zap.Error(err), zap.Int("status", status), zap.Any("request_id", requestID))
		c.AbortWithStatusJSON(status, entity.ErrorResponse{Code: int32(status), Message: message})
	}
}
