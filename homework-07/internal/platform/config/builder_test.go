package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type customConfig struct {
	Value string `mapstructure:"value" validate:"required"`
}

func TestBuilderAppliesCustomDefault(t *testing.T) {
	cfg, err := NewBuilder[customConfig]().
		WithDefaultValues(map[string]any{"value": "default-value"}).
		Load()

	require.NoError(t, err)
	require.Equal(t, "default-value", cfg.Value)
}

func TestBuilderPrefixesCustomEnvironmentBinding(t *testing.T) {
	t.Setenv("TEST_SERVICE_CUSTOM_VALUE", "environment-value")
	cfg, err := NewBuilder[customConfig]().
		WithEnvPrefix("TEST_SERVICE").
		WithBindEnv(map[string]string{"value": "CUSTOM_VALUE"}).
		WithDefaultValues(map[string]any{"value": "default-value"}).
		Load()

	require.NoError(t, err)
	require.Equal(t, "environment-value", cfg.Value)
}
