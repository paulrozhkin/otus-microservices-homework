package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-08-09/internal/platform/apperror"
)

type BillingClient struct {
	baseURL string
	client  *http.Client
}

type paymentRequest struct {
	OperationID string `json:"operationId"`
	UserID      int64  `json:"userId"`
	Amount      int64  `json:"amount"`
}

func NewBillingClient(baseURL string, timeout time.Duration) *BillingClient {
	return &BillingClient{baseURL: strings.TrimRight(baseURL, "/"), client: &http.Client{Timeout: timeout}}
}

func (c *BillingClient) Pay(ctx context.Context, orderID string, userID, amount int64) error {
	payload, err := json.Marshal(paymentRequest{OperationID: "order:" + orderID, UserID: userID, Amount: amount})
	if err != nil {
		return fmt.Errorf("marshal payment request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/payments", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create payment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("execute payment request: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return apperror.ErrInsufficientFunds
	case http.StatusNotFound:
		return apperror.ErrNotFound
	default:
		return fmt.Errorf("billing payment returned unexpected status %d", resp.StatusCode)
	}
}
