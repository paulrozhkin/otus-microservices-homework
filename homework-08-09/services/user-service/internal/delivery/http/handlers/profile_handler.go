package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/user-service/internal/repositories"
	"go.uber.org/zap"
)

const (
	authUserIDHeader = "X-Auth-UserId"
	authRolesHeader  = "X-Auth-Roles"
)

type ProfileHandler struct {
	userRepository repositories.UserRepository
	logger         *zap.Logger
}

type UpdateProfileRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone" binding:"required"`
}

func NewProfileHandler(logger *zap.Logger, userRepository repositories.UserRepository) *ProfileHandler {
	return &ProfileHandler{userRepository: userRepository, logger: logger}
}

func (h *ProfileHandler) Get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.Error(apperror.ErrUnauthorized)
		return
	}

	user, err := h.userRepository.GetUserByID(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *ProfileHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.Error(apperror.ErrUnauthorized)
		return
	}

	currentUser, err := h.userRepository.GetUserByID(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	request := &UpdateProfileRequest{}
	if err = c.ShouldBindJSON(request); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}

	currentUser.FirstName = request.FirstName
	currentUser.LastName = request.LastName
	currentUser.Email = request.Email
	currentUser.Phone = request.Phone

	user, err := h.userRepository.UpdateUser(c, currentUser)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func currentUserID(c *gin.Context) (int64, bool) {
	value := c.GetHeader(authUserIDHeader)
	if value == "" {
		return 0, false
	}

	userID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false
	}

	return userID, true
}
