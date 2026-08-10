package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/notification-service/internal/repositories"
)

const authUserIDHeader = "X-Auth-UserId"

type NotificationHandler struct {
	repository repositories.NotificationRepository
}

func NewNotificationHandler(repository repositories.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{repository: repository}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, err := strconv.ParseInt(c.GetHeader(authUserIDHeader), 10, 64)
	if err != nil || userID <= 0 {
		c.Error(apperror.ErrUnauthorized)
		return
	}
	notifications, err := h.repository.List(c, userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, notifications)
}
