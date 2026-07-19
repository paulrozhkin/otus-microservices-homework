package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type accountProvisionerStub struct {
	userID int64
}

func (s *accountProvisionerStub) CreateAccount(_ context.Context, userID int64) error {
	s.userID = userID
	return nil
}

func TestRegisterCreatesBillingAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockUserRepository(ctrl)
	repository.EXPECT().CreateUser(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, user *entity.User) (*entity.User, error) {
			user.Id = 42
			return user, nil
		},
	)
	provisioner := &accountProvisionerStub{}
	handler := NewAuthHandler(zap.NewNop(), repository, provisioner)
	router := gin.New()
	router.POST("/register", handler.Register)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{
		"login":"user-1",
		"password":"secret",
		"email":"user-1@example.com",
		"firstName":"Test",
		"lastName":"User",
		"phone":"+70000000000"
	}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	require.Equal(t, int64(42), provisioner.userID)
}
