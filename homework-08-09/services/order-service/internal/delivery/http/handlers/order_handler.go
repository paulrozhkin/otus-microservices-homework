package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/service"
)

const (
	authUserIDHeader = "X-Auth-UserId"
	authEmailHeader  = "X-Auth-Email"
)

type OrderHandler struct{ service *service.OrderService }

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

type CreateOrderRequest struct {
	Price int64 `json:"price" binding:"required,gt=0"`
}

func (h *OrderHandler) Create(c *gin.Context) {
	userID, email, ok := authContext(c)
	if !ok {
		return
	}
	request := &CreateOrderRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	order, err := h.service.Create(c, userID, email, request.Price)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) Get(c *gin.Context) {
	userID, _, ok := authContext(c)
	if !ok {
		return
	}
	order, err := h.service.Get(c, c.Param("id"), userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	userID, _, ok := authContext(c)
	if !ok {
		return
	}
	orders, err := h.service.List(c, userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, orders)
}

func authContext(c *gin.Context) (int64, string, bool) {
	userID, err := strconv.ParseInt(c.GetHeader(authUserIDHeader), 10, 64)
	email := c.GetHeader(authEmailHeader)
	if err != nil || userID <= 0 || email == "" {
		c.Error(apperror.ErrUnauthorized)
		return 0, "", false
	}
	return userID, email, true
}
