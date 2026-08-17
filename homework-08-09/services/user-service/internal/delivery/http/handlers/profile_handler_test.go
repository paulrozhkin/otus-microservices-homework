package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/user-service/internal/entity"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/user-service/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestProfileHandlerGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns current user profile", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockUserRepository(ctrl)
		repo.EXPECT().
			GetUserByID(gomock.Any(), int64(42)).
			Return(&entity.User{
				Id:        42,
				Username:  "paul",
				FirstName: "Paul",
				LastName:  "Rozhkin",
				Email:     "paul@example.com",
				Phone:     "+71002003040",
			}, nil)

		handler := NewProfileHandler(zap.NewNop(), repo)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)
		c.Request.Header.Set(authUserIDHeader, "42")

		handler.Get(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.Empty(t, c.Errors)
		require.JSONEq(t, `{
			"id": 42,
			"username": "paul",
			"firstName": "Paul",
			"lastName": "Rozhkin",
			"email": "paul@example.com",
			"phone": "+71002003040"
		}`, w.Body.String())
	})

	t.Run("requires auth header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repo := mocks.NewMockUserRepository(ctrl)

		handler := NewProfileHandler(zap.NewNop(), repo)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/profile", nil)

		handler.Get(c)

		require.Equal(t, http.StatusOK, w.Code)
		require.Len(t, c.Errors, 1)
		require.ErrorIs(t, c.Errors.Last().Err, apperror.ErrUnauthorized)
	})
}

func TestProfileHandlerUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	repo.EXPECT().
		GetUserByID(gomock.Any(), int64(42)).
		Return(&entity.User{
			Id:           42,
			Username:     "paul",
			PasswordHash: "hash",
			FirstName:    "Old",
			LastName:     "Name",
			Email:        "old@example.com",
			Phone:        "+70000000000",
			Roles:        []string{"user"},
		}, nil)
	repo.EXPECT().
		UpdateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user *entity.User) (*entity.User, error) {
			require.Equal(t, int64(42), user.Id)
			require.Equal(t, "paul", user.Username)
			require.Equal(t, "hash", user.PasswordHash)
			require.Equal(t, []string{"user"}, user.Roles)
			require.Equal(t, "Paul", user.FirstName)
			require.Equal(t, "Rozhkin", user.LastName)
			require.Equal(t, "paul@example.com", user.Email)
			require.Equal(t, "+71002003040", user.Phone)
			return user, nil
		})

	handler := NewProfileHandler(zap.NewNop(), repo)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/profile", bytes.NewBufferString(`{
		"firstName": "Paul",
		"lastName": "Rozhkin",
		"email": "paul@example.com",
		"phone": "+71002003040"
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set(authUserIDHeader, "42")

	handler.Update(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Empty(t, c.Errors)
	require.JSONEq(t, `{
		"id": 42,
		"username": "paul",
		"firstName": "Paul",
		"lastName": "Rozhkin",
		"email": "paul@example.com",
		"phone": "+71002003040",
		"roles": ["user"]
	}`, w.Body.String())
}
