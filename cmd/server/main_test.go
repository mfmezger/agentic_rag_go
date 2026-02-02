package main

import (
	"os"
	"testing"

	"github.com/mfmezger/agentic_rag_go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_Configuration(t *testing.T) {
	tests := []struct {
		name       string
		envKey     string
		configPath string
		wantErr    bool
	}{
		{
			name:       "valid config with env var",
			envKey:     "test-api-key",
			configPath: "",
			wantErr:    false,
		},
		{
			name:       "valid config with env var and special env vars",
			envKey:     "test-api-key",
			configPath: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				os.Setenv("GOOGLE_API_KEY", tt.envKey)
				defer os.Unsetenv("GOOGLE_API_KEY")
			}

			cfg, err := config.Load(tt.configPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
			}
		})
	}
}

func TestMain_ConfigDefaults(t *testing.T) {
	os.Unsetenv("GOOGLE_API_KEY")

	cfg, err := config.Load("")
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8001, cfg.Server.Port)
	assert.Equal(t, "qdrant", cfg.VectorStore.Provider)
}

func TestMain_InvalidConfig(t *testing.T) {
	os.Setenv("GOOGLE_API_KEY", "")
	defer os.Unsetenv("GOOGLE_API_KEY")

	os.Setenv("CONFIG_PATH", "/nonexistent/config.yaml")
	defer os.Unsetenv("CONFIG_PATH")

	cfg, err := config.Load("/nonexistent/config.yaml")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestMain_EnvFileLoading(t *testing.T) {
	tests := []struct {
		name       string
		envFile    string
		wantConfig bool
	}{
		{
			name:       "no env file",
			envFile:    "",
			wantConfig: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envFile != "" {
				os.Setenv("CONFIG_PATH", tt.envFile)
				defer os.Unsetenv("CONFIG_PATH")
			}

			os.Unsetenv("GOOGLE_API_KEY")

			cfg, err := config.Load("")
			require.NoError(t, err)
			assert.NotNil(t, cfg)
		})
	}
}

func TestMain_SpecialEnvVars_QdrantURL(t *testing.T) {
	os.Unsetenv("GOOGLE_API_KEY")

	t.Setenv("QDRANT_URL", "qdrant-special.com")

	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.Equal(t, "qdrant-special.com", cfg.VectorStore.URL)
}

func TestMain_SpecialEnvVars_Phoenix(t *testing.T) {
	os.Unsetenv("GOOGLE_API_KEY")

	t.Setenv("PHOENIX_COLLECTOR_ENDPOINT", "http://phoenix:6006")

	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.Equal(t, "http://phoenix:6006", cfg.Tracing.Endpoint)
	assert.True(t, cfg.Tracing.Enabled)
}

func TestMain_SpecialEnvVars_OTEL(t *testing.T) {
	os.Unsetenv("GOOGLE_API_KEY")

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel:4317")

	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.Equal(t, "http://otel:4317", cfg.Tracing.Endpoint)
	assert.True(t, cfg.Tracing.Enabled)
}

func TestMain_Precedence_EnvOverYAML(t *testing.T) {
	os.Unsetenv("GOOGLE_API_KEY")

	t.Setenv("APP_MODEL_NAME", "env-override-model")
	t.Setenv("APP_VECTORSTORE_COLLECTION", "env_collection")

	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.Equal(t, "env-override-model", cfg.Model.Name)
	assert.Equal(t, "env_collection", cfg.VectorStore.Collection)
}

func TestMain_Precedence_APP_EnvOverSpecialEnv(t *testing.T) {
	os.Unsetenv("GOOGLE_API_KEY")

	t.Setenv("APP_VECTORSTORE_URL", "app-url.com")
	t.Setenv("QDRANT_URL", "qdrant-url.com")

	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.Equal(t, "app-url.com", cfg.VectorStore.URL)
}

func TestMain_ConfigServerFields(t *testing.T) {
	cfg, err := config.Load("")
	require.NoError(t, err)

	assert.Equal(t, "", cfg.Server.APIKey)
	assert.Equal(t, 100, cfg.Server.RateLimit)
	assert.Equal(t, 60, cfg.Server.RateWindow)
}
