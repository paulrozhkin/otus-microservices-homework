package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-05/internal/entity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestErrorHandlerMapsErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		errType    gin.ErrorType
		wantStatus int
		wantBody   string
	}{
		{
			name:       "bind error",
			err:        errors.New("invalid request"),
			errType:    gin.ErrorTypeBind,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"code":400,"message":"invalid request"}`,
		},
		{
			name:       "already exists",
			err:        entity.ErrAlreadyExists,
			wantStatus: http.StatusConflict,
			wantBody:   `{"code":409,"message":"already exists"}`,
		},
		{
			name:       "not found",
			err:        entity.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   `{"code":404,"message":"not found"}`,
		},
		{
			name:       "service unavailable",
			err:        entity.ErrServiceUnavailable,
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"code":503,"message":"service unavailable"}`,
		},
		{
			name:       "unknown error",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"code":500,"message":"Internal Server Error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestID())
			router.Use(ErrorHandler(zap.NewNop()))
			router.GET("/test", func(c *gin.Context) {
				ginErr := c.Error(tt.err)
				if tt.errType != 0 {
					ginErr.SetType(tt.errType)
				}
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			router.ServeHTTP(w, req)

			require.Equal(t, tt.wantStatus, w.Code)
			require.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestErrorHandlerDoesNothingWithoutErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler(zap.NewNop()))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)
	require.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}
