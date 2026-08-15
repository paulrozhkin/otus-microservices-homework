package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	platformconfig "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/order-service/internal/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOrderSwagger(t *testing.T) {
	router := NewRouter(RouterConfig{
		Config: config.Config{BaseConfig: platformconfig.BaseConfig{App: platformconfig.AppConfig{Env: platformconfig.DevelopmentEnv}}},
		Logger: zap.NewNop(),
	})

	t.Run("yaml", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil))
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Header().Get("Content-Type"), "application/yaml")
		require.Contains(t, w.Body.String(), "title: Order Service")
		require.Contains(t, w.Body.String(), "/api/v1/orders:")
	})

	t.Run("ui", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger", nil))
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Header().Get("Content-Type"), "text/html")
		require.Contains(t, w.Body.String(), "Order Service Swagger")
	})
}
