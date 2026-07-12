package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-05/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-05/internal/repositories"
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
		c.Error(err)
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
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	createdUser, err := h.userRepository.CreateUser(c, request)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	var params RequestParams

	// Binds the path parameters to the struct
	if err := c.ShouldBindUri(&params); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
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
	err := h.userRepository.DeleteUser(c, params.Id)
	if err != nil {
		c.Error(err)
		return
	}
	c.Status(http.StatusOK)
}
