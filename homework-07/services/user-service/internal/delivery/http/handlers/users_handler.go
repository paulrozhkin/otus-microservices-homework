package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/entity"
	businessmetrics "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/metrics"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/repositories"
	"go.uber.org/zap"
)

type UserHandler struct {
	userRepository     repositories.UserRepository
	accountProvisioner AccountProvisioner
	logger             *zap.Logger
}

func NewUserHandler(logger *zap.Logger, userRepository repositories.UserRepository, provisioners ...AccountProvisioner) *UserHandler {
	handler := &UserHandler{userRepository: userRepository, logger: logger}
	if len(provisioners) > 0 {
		handler.accountProvisioner = provisioners[0]
	}
	return handler
}

func (h *UserHandler) GetAll(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	users, err := h.userRepository.GetAllUsers(c)
	if err != nil {
		c.Error(err)
		return
	}
	if len(users) == 0 {
		users = []*entity.User{}
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Create(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}

	request := &entity.User{}
	err := c.ShouldBindJSON(request)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	createdUser, err := h.userRepository.CreateUser(c, request)
	if err != nil {
		c.Error(err)
		return
	}
	if h.accountProvisioner != nil {
		if err = h.accountProvisioner.CreateAccount(c, createdUser.Id); err != nil {
			c.Error(err)
			return
		}
	}
	businessmetrics.UserCreated("admin_api")

	c.JSON(http.StatusCreated, createdUser)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	var params RequestParams

	// Binds the path parameters to the struct
	if err := c.ShouldBindUri(&params); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	if !requireAdminOrSameUser(c, params.Id) {
		return
	}

	user, err := h.userRepository.GetUserByID(c, params.Id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	var params RequestParams

	// Binds the path parameters to the struct
	if err := c.ShouldBindUri(&params); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	if !requireAdminOrSameUser(c, params.Id) {
		return
	}

	request := &entity.User{}
	err := c.ShouldBindJSON(request)
	if err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	request.Id = params.Id
	user, err := h.userRepository.UpdateUser(c, request)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	var params RequestParams

	// Binds the path parameters to the struct
	if err := c.ShouldBindUri(&params); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	if !requireAdmin(c) {
		return
	}

	err := h.userRepository.DeleteUser(c, params.Id)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}

func requireAdmin(c *gin.Context) bool {
	if !hasAuthContext(c) {
		c.Error(apperror.ErrUnauthorized)
		return false
	}
	if hasRole(c, "admin") {
		return true
	}
	c.Error(apperror.ErrForbidden)
	return false
}

func requireAdminOrSameUser(c *gin.Context, targetUserID int64) bool {
	currentUserID, ok := currentUserID(c)
	if !ok || c.GetHeader(authRolesHeader) == "" {
		c.Error(apperror.ErrUnauthorized)
		return false
	}
	if hasRole(c, "admin") {
		return true
	}
	if currentUserID == targetUserID {
		return true
	}
	c.Error(apperror.ErrForbidden)
	return false
}

func hasAuthContext(c *gin.Context) bool {
	_, ok := currentUserID(c)
	return ok && c.GetHeader(authRolesHeader) != ""
}

func hasRole(c *gin.Context, expected string) bool {
	for _, role := range strings.Split(c.GetHeader(authRolesHeader), ",") {
		if strings.EqualFold(strings.TrimSpace(role), expected) {
			return true
		}
	}
	return false
}
