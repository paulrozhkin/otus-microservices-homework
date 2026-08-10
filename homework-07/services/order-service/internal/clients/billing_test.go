package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/internal/platform/apperror"
	"github.com/stretchr/testify/require"
)

func TestBillingClientSendsIdempotentPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/internal/v1/payments", r.URL.Path)
		var request paymentRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, "order:order-1", request.OperationID)
		require.Equal(t, int64(42), request.UserID)
		require.Equal(t, int64(10000), request.Amount)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := NewBillingClient(server.URL, time.Second).Pay(t.Context(), "order-1", 42, 10000)
	require.NoError(t, err)
}

func TestBillingClientMapsRejectedPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusConflict) }))
	defer server.Close()

	err := NewBillingClient(server.URL, time.Second).Pay(t.Context(), "order-1", 42, 10000)
	require.ErrorIs(t, err, apperror.ErrInsufficientFunds)
}
