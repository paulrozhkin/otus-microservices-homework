package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/repositories"
)

const authUserIDHeader = "X-Auth-UserId"

type BillingHandler struct {
	repository repositories.BillingRepository
}

func NewBillingHandler(repository repositories.BillingRepository) *BillingHandler {
	return &BillingHandler{repository: repository}
}

type AmountRequest struct {
	Amount      int64  `json:"amount" binding:"required,gt=0"`
	OperationID string `json:"operationId"`
}

type PaymentRequest struct {
	OperationID string `json:"operationId" binding:"required,max=128"`
	UserID      int64  `json:"userId" binding:"required,gt=0"`
	Amount      int64  `json:"amount" binding:"required,gt=0"`
}

func (h *BillingHandler) CreateAccount(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil || userID <= 0 {
		c.Error(entity.ErrInvalidOperation).SetType(gin.ErrorTypeBind)
		return
	}
	account, created, err := h.repository.CreateAccount(c, userID)
	if err != nil {
		c.Error(err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, account)
}

func (h *BillingHandler) GetAccount(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	account, err := h.repository.GetAccount(c, userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *BillingHandler) Deposit(c *gin.Context) { h.changeBalance(c, false, entity.OperationDeposit) }
func (h *BillingHandler) Withdraw(c *gin.Context) {
	h.changeBalance(c, true, entity.OperationWithdrawal)
}

func (h *BillingHandler) changeBalance(c *gin.Context, debit bool, operationType string) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	request := &AmountRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	if request.OperationID == "" {
		request.OperationID = uuid.NewString()
	}
	var account *entity.Account
	var err error
	if debit {
		account, err = h.repository.Debit(c, request.OperationID, userID, request.Amount, operationType)
	} else {
		account, err = h.repository.Credit(c, request.OperationID, userID, request.Amount, operationType)
	}
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"operationId": request.OperationID, "account": account})
}

func (h *BillingHandler) Pay(c *gin.Context) {
	request := &PaymentRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		c.Error(err).SetType(gin.ErrorTypeBind)
		return
	}
	account, err := h.repository.Debit(c, request.OperationID, request.UserID, request.Amount, entity.OperationPayment)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"operationId": request.OperationID, "status": "succeeded", "account": account})
}

func (h *BillingHandler) Refund(c *gin.Context) {
	account, err := h.repository.Refund(c, c.Param("operationId"))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"operationId": c.Param("operationId"), "status": "refunded", "account": account})
}

func authenticatedUserID(c *gin.Context) (int64, bool) {
	userID, err := strconv.ParseInt(c.GetHeader(authUserIDHeader), 10, 64)
	if err != nil || userID <= 0 {
		c.Error(entity.ErrUnauthorized)
		return 0, false
	}
	return userID, true
}
