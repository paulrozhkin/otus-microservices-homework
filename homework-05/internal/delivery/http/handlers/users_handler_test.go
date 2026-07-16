package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-05/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-05/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestUserHandlerGetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMock      func(*mocks.MockUserRepository)
		wantStatus     int
		wantBody       string
		wantErrorCount int
	}{
		{
			name: "returns users",
			setupMock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					GetAllUsers(gomock.Any()).
					Return([]*entity.User{{Id: 1, Username: "paul"}}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `[{"id":1,"username":"paul","firstName":"","lastName":"","email":"","phone":""}]`,
		},
		{
			name: "returns empty array",
			setupMock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					GetAllUsers(gomock.Any()).
					Return(nil, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name: "records repository error",
			setupMock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().
					GetAllUsers(gomock.Any()).
					Return(nil, errors.New("db failed"))
			},
			wantStatus:     http.StatusOK,
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockUserRepository(ctrl)
			tt.setupMock(repo)

			handler := NewUserHandler(zap.NewNop(), repo)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)

			handler.GetAll(c)

			require.Equal(t, tt.wantStatus, w.Code)
			require.Len(t, c.Errors, tt.wantErrorCount)
			if tt.wantBody != "" {
				require.JSONEq(t, tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestUserHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockUserRepository(ctrl)
		repo.EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, user *entity.User) (*entity.User, error) {
				require.Equal(t, "paul", user.Username)
				user.Id = 10
				return user, nil
			})

		handler := NewUserHandler(zap.NewNop(), repo)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{
			"username": "paul",
			"firstName": "Paul",
			"lastName": "Rozhkin",
			"email": "paul@example.com",
			"phone": "+71002003040"
		}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		require.Equal(t, http.StatusCreated, w.Code)
		require.JSONEq(t, `{
			"id": 10,
			"username": "paul",
			"firstName": "Paul",
			"lastName": "Rozhkin",
			"email": "paul@example.com",
			"phone": "+71002003040"
		}`, w.Body.String())
		require.Empty(t, c.Errors)
	})

	t.Run("records bind error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockUserRepository(ctrl)

		handler := NewUserHandler(zap.NewNop(), repo)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"username": ""}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.Len(t, c.Errors, 1)
		require.Equal(t, gin.ErrorTypeBind, c.Errors.Last().Type)
	})
}

func TestUserHandlerGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&entity.User{Id: 42, Username: "paul"}, nil)

	handler := NewUserHandler(zap.NewNop(), repo)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users/42", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}}

	handler.GetByID(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"id":42,"username":"paul","firstName":"","lastName":"","email":"","phone":""}`, w.Body.String())
}

func TestUserHandlerUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	repo.EXPECT().
		UpdateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user *entity.User) (*entity.User, error) {
			require.Equal(t, int64(42), user.Id)
			return user, nil
		})

	handler := NewUserHandler(zap.NewNop(), repo)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/users/42", bytes.NewBufferString(`{
		"username": "paul",
		"firstName": "Paul",
		"lastName": "Rozhkin",
		"email": "paul@example.com",
		"phone": "+71002003040"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "42"}}

	handler.Update(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, c.Errors)
}

func TestUserHandlerDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	repo.EXPECT().
		DeleteUser(gomock.Any(), int64(42)).
		Return(nil)

	handler := NewUserHandler(zap.NewNop(), repo)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/users/42", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}}

	handler.Delete(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, c.Errors)
}
