package repositories

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestIsUniqueViolation(t *testing.T) {
	require.True(t, isUniqueViolation(&pgconn.PgError{Code: uniqueViolationCode}))
	require.True(t, isUniqueViolation(errors.Join(errors.New("wrapped"), &pgconn.PgError{Code: uniqueViolationCode})))
	require.False(t, isUniqueViolation(&pgconn.PgError{Code: "23503"}))
	require.False(t, isUniqueViolation(errors.New("boom")))
}

func TestIsNotFoundViolation(t *testing.T) {
	require.True(t, isNotFoundViolation(gorm.ErrRecordNotFound))
	require.True(t, isNotFoundViolation(errors.Join(errors.New("wrapped"), gorm.ErrRecordNotFound)))
	require.False(t, isNotFoundViolation(errors.New("boom")))
}
