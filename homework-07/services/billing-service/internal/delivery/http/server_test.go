package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/billing-service/internal/entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeBillingRepository struct {
	account  *entity.Account
	debitErr error
}

func (f *fakeBillingRepository) CreateAccount(_ context.Context, userID int64) (*entity.Account, bool, error) {
	f.account = &entity.Account{UserID: userID}
	return f.account, true, nil
}
func (f *fakeBillingRepository) GetAccount(context.Context, int64) (*entity.Account, error) {
	return f.account, nil
}
func (f *fakeBillingRepository) Credit(_ context.Context, _ string, userID, amount int64, _ string) (*entity.Account, error) {
	f.account = &entity.Account{UserID: userID, Balance: amount}
	return f.account, nil
}
func (f *fakeBillingRepository) Debit(context.Context, string, int64, int64, string) (*entity.Account, error) {
	return f.account, f.debitErr
}
func (f *fakeBillingRepository) Refund(context.Context, string) (*entity.Account, error) {
	return f.account, nil
}

func newTestRouter(repository *fakeBillingRepository) http.Handler {
	gin.SetMode(gin.TestMode)
	return NewRouter(RouterConfig{
		Config: config.Config{App: config.AppConfig{Env: config.DevelopmentEnv}},
		Logger: zap.NewNop(), Repository: repository,
	})
}

func TestCreateAccount(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/internal/v1/accounts/42", nil)
	newTestRouter(&fakeBillingRepository{}).ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	require.Contains(t, w.Body.String(), `"userId":42`)
	require.Contains(t, w.Body.String(), `"balance":0`)
}

func TestDepositUsesAuthenticatedUser(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/deposits", strings.NewReader(`{"amount":10000,"operationId":"deposit-1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-UserId", "42")
	newTestRouter(&fakeBillingRepository{}).ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"balance":10000`)
}

func TestPaymentRejectsInsufficientFunds(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/payments", strings.NewReader(`{"operationId":"order-1","userId":42,"amount":10000}`))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter(&fakeBillingRepository{debitErr: errors.Join(errors.New("wrapped"), entity.ErrInsufficientFunds)}).ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
	require.Contains(t, w.Body.String(), "insufficient funds")
}

func TestBillingAPIRequiresAuthHeader(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/billing/account", nil)
	newTestRouter(&fakeBillingRepository{}).ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
