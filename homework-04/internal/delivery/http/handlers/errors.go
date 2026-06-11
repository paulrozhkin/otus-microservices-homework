package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/entity"
	"go.uber.org/zap"
)

func ResponseError(c *gin.Context, logger *zap.Logger, code int, err error) {
	logger.Error("failed to execute handler", zap.Error(err))
	errorMessage := entity.ErrorResponse{Message: err.Error(), Code: int32(code)}
	c.JSON(code, errorMessage)
}
