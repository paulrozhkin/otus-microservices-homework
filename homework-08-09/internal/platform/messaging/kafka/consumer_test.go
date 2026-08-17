package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWaitForRetryStopsWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	startedAt := time.Now()
	err := waitForRetry(ctx, time.Minute)

	require.ErrorIs(t, err, context.Canceled)
	require.Less(t, time.Since(startedAt), time.Second)
}

func TestWaitForRetryWaitsForDelay(t *testing.T) {
	startedAt := time.Now()

	require.NoError(t, waitForRetry(context.Background(), 10*time.Millisecond))
	require.GreaterOrEqual(t, time.Since(startedAt), 10*time.Millisecond)
}
