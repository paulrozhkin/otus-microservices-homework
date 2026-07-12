package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-05/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-05/internal/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestRouterLiveness(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(RouterConfig{
		Config: config.Config{App: config.AppConfig{Env: config.DevelopmentEnv}},
		Logger: zap.NewNop(),
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestRouterReadiness(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "ready",
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok"}`,
		},
		{
			name:       "unavailable",
			pingErr:    errors.New("db down"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"code":503,"message":"service unavailable"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			healthChecker := mocks.NewMockHealthChecker(ctrl)
			healthChecker.EXPECT().
				Ping(gomock.Any()).
				Return(tt.pingErr)

			router := NewRouter(RouterConfig{
				Config:        config.Config{App: config.AppConfig{Env: config.DevelopmentEnv}},
				Logger:        zap.NewNop(),
				HealthChecker: healthChecker,
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

			router.ServeHTTP(w, req)

			require.Equal(t, tt.wantStatus, w.Code)
			require.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestRouterSwagger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(RouterConfig{
		Config: config.Config{App: config.AppConfig{Env: config.DevelopmentEnv}},
		Logger: zap.NewNop(),
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "title: User Service")
	require.Contains(t, w.Body.String(), "/users:")
}
