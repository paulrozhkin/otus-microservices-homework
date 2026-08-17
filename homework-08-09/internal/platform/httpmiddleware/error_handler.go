package httpmiddleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"go.uber.org/zap"
)

type ErrorMapping struct {
	Target error
	Status int
}

var defaultErrorMappings = []ErrorMapping{
	{Target: apperror.ErrInvalidOperation, Status: http.StatusBadRequest},
	{Target: apperror.ErrUnauthorized, Status: http.StatusUnauthorized},
	{Target: apperror.ErrForbidden, Status: http.StatusForbidden},
	{Target: apperror.ErrNotFound, Status: http.StatusNotFound},
	{Target: apperror.ErrAlreadyExists, Status: http.StatusConflict},
	{Target: apperror.ErrInsufficientFunds, Status: http.StatusConflict},
	{Target: apperror.ErrServiceUnavailable, Status: http.StatusServiceUnavailable},
}

// ErrorHandler maps common application errors to HTTP responses. Service-specific
// mappings can be appended without creating another middleware implementation.
func ErrorHandler(logger *zap.Logger, mappings ...ErrorMapping) gin.HandlerFunc {
	allMappings := append(append([]ErrorMapping{}, mappings...), defaultErrorMappings...)
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}

		ginErr := c.Errors.Last()
		err := ginErr.Err
		status := http.StatusInternalServerError
		message := http.StatusText(status)
		if ginErr.Type == gin.ErrorTypeBind {
			status, message = http.StatusBadRequest, err.Error()
		} else {
			for _, mapping := range allMappings {
				if errors.Is(err, mapping.Target) {
					status, message = mapping.Status, err.Error()
					break
				}
			}
		}

		requestID, _ := c.Get(RequestIDKey)
		logger.Error("request failed", zap.Error(err), zap.Int("status", status), zap.Any("request_id", requestID))
		c.AbortWithStatusJSON(status, apperror.Response{Code: int32(status), Message: message})
	}
}
