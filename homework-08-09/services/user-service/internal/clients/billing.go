package clients

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type BillingClient struct {
	baseURL string
	client  *http.Client
}

func NewBillingClient(baseURL string, timeout time.Duration) *BillingClient {
	return &BillingClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *BillingClient) CreateAccount(ctx context.Context, userID int64) error {
	url := fmt.Sprintf("%s/internal/v1/accounts/%d", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	if err != nil {
		return fmt.Errorf("create billing request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("create billing account for user %d: %w", userID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create billing account for user %d: unexpected status %d", userID, resp.StatusCode)
	}

	return nil
}
