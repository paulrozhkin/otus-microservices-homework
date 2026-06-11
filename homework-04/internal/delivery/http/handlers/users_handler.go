package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-04/internal/repositories"
	"go.uber.org/zap"
)

type UserHandler struct {
	userRepository repositories.UserRepository
	logger         *zap.Logger
}

func NewUserHandler(logger *zap.Logger, userRepository repositories.UserRepository) *UserHandler {
	return &UserHandler{userRepository: userRepository, logger: logger}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userRepository.GetAllUsers(c)
	if err != nil {
		ResponseError(c, h.logger, http.StatusInternalServerError, err)
		return
	}
	if len(users) == 0 {
		users = []*entity.User{}
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Create(c *gin.Context) {
	request := &entity.User{}
	err := c.ShouldBindJSON(request)
	if err != nil {
		ResponseError(c, h.logger, http.StatusBadRequest, err)
		return
	}
	createdUser, err := h.userRepository.CreateUser(c, request)
	if err != nil {
		ResponseError(c, h.logger, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, createdUser)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	c.Error(fmt.Errorf("method is not implemented"))
}

func (h *UserHandler) Update(c *gin.Context) {
	c.Error(fmt.Errorf("method is not implemented"))
}

func (h *UserHandler) Delete(c *gin.Context) {
	c.Error(fmt.Errorf("method is not implemented"))
}
