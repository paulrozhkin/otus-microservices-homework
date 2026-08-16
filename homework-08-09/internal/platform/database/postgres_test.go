package database

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOpenPostgresWithRetryEventuallySucceeds(t *testing.T) {
	attempts := 0
	expected := &gorm.DB{}
	db, err := openPostgresWithRetry("secret-dsn", "postgres:5432", 100*time.Millisecond, time.Millisecond, func(string) (*gorm.DB, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("not ready")
		}
		return expected, nil
	})

	require.NoError(t, err)
	require.Same(t, expected, db)
	require.Equal(t, 3, attempts)
}

func TestOpenPostgresWithRetryStopsAfterTimeoutWithoutLeakingDSN(t *testing.T) {
	attempts := 0
	_, err := openPostgresWithRetry("password=super-secret", "postgres:5432", 10*time.Millisecond, 2*time.Millisecond, func(string) (*gorm.DB, error) {
		attempts++
		return nil, errors.New("connection refused")
	})

	require.Error(t, err)
	require.GreaterOrEqual(t, attempts, 2)
	require.Contains(t, err.Error(), "postgres:5432")
	require.Contains(t, err.Error(), "connection refused")
	require.False(t, strings.Contains(err.Error(), "super-secret"))
}
