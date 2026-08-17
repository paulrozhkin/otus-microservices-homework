package httpmiddleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/apperror"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestErrorHandlerMapsErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		errType    gin.ErrorType
		wantStatus int
		wantBody   string
	}{
		{name: "bind", err: errors.New("invalid request"), errType: gin.ErrorTypeBind, wantStatus: http.StatusBadRequest, wantBody: `{"code":400,"message":"invalid request"}`},
		{name: "invalid operation", err: apperror.ErrInvalidOperation, wantStatus: http.StatusBadRequest, wantBody: `{"code":400,"message":"invalid operation"}`},
		{name: "unauthorized", err: apperror.ErrUnauthorized, wantStatus: http.StatusUnauthorized, wantBody: `{"code":401,"message":"unauthorized"}`},
		{name: "forbidden", err: apperror.ErrForbidden, wantStatus: http.StatusForbidden, wantBody: `{"code":403,"message":"forbidden"}`},
		{name: "not found", err: apperror.ErrNotFound, wantStatus: http.StatusNotFound, wantBody: `{"code":404,"message":"not found"}`},
		{name: "already exists", err: apperror.ErrAlreadyExists, wantStatus: http.StatusConflict, wantBody: `{"code":409,"message":"already exists"}`},
		{name: "insufficient funds", err: apperror.ErrInsufficientFunds, wantStatus: http.StatusConflict, wantBody: `{"code":409,"message":"insufficient funds"}`},
		{name: "unavailable", err: apperror.ErrServiceUnavailable, wantStatus: http.StatusServiceUnavailable, wantBody: `{"code":503,"message":"service unavailable"}`},
		{name: "unknown", err: errors.New("boom"), wantStatus: http.StatusInternalServerError, wantBody: `{"code":500,"message":"Internal Server Error"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(RequestID(), ErrorHandler(zap.NewNop()))
			router.GET("/test", func(c *gin.Context) {
				ginErr := c.Error(tt.err)
				if tt.errType != 0 {
					ginErr.SetType(tt.errType)
				}
			})
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
			require.Equal(t, tt.wantStatus, w.Code)
			require.JSONEq(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestErrorHandlerAllowsSuccessfulResponse(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandler(zap.NewNop()))
	router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusAccepted, gin.H{"status": "ok"}) })
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	require.Equal(t, http.StatusAccepted, w.Code)
}
