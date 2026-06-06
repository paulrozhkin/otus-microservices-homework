package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) Create(c *gin.Context) {
	c.Error(fmt.Errorf("method is not implemented"))
}

func (h *UserHandler) GetByID(c *gin.Context) {
	c.Error(fmt.Errorf("method is not implemented"))
}

func (h *UserHandler) List(c *gin.Context) {
	c.Error(fmt.Errorf("method is not implemented"))
}

func (h *UserHandler) Update(c *gin.Context) {
	c.Error(fmt.Errorf("method is not implemented"))
}

func (h *UserHandler) Delete(c *gin.Context) {
	c.Error(fmt.Errorf("method is not implemented"))
}
