package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	platformconfig "github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/config"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/services/notification-service/internal/entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeNotificationRepository struct {
	userID        int64
	notifications []*entity.Notification
}

func (f *fakeNotificationRepository) Create(context.Context, *entity.Notification) error { return nil }

func (f *fakeNotificationRepository) List(_ context.Context, userID int64) ([]*entity.Notification, error) {
	f.userID = userID
	return f.notifications, nil
}

func newTestRouter(repository *fakeNotificationRepository) http.Handler {
	return NewRouter(RouterConfig{
		Config: config.Config{BaseConfig: platformconfig.BaseConfig{App: platformconfig.AppConfig{Env: platformconfig.DevelopmentEnv}}},
		Logger: zap.NewNop(), Repository: repository,
	})
}

func TestListNotificationsUsesAuthenticatedUser(t *testing.T) {
	repository := &fakeNotificationRepository{notifications: []*entity.Notification{{
		EventID: "event-1", OrderID: "order-1", UserID: 42, Email: "user@example.com",
		OrderStatus: "paid", Subject: "Order paid", Body: "Success", CreatedAt: time.Now().UTC(),
	}}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.Header.Set("X-Auth-UserId", "42")

	newTestRouter(repository).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, int64(42), repository.userID)
	require.Contains(t, w.Body.String(), `"eventId":"event-1"`)
}

func TestNotificationsRequireAuthHeader(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)

	newTestRouter(&fakeNotificationRepository{}).ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNotificationSwagger(t *testing.T) {
	router := newTestRouter(&fakeNotificationRepository{})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil))
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "title: Notification Service")
	require.Contains(t, w.Body.String(), "/api/v1/notifications:")
}
