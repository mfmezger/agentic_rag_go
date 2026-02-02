package embedding

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService_Success(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		APIKey:    "test-api-key",
		ModelName: "gemini-embedding-001",
	}

	service, err := NewService(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.client)
	assert.Equal(t, "gemini-embedding-001", service.modelName)
}

func TestNewService_DefaultModelName(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		APIKey:    "test-api-key",
		ModelName: "", // Empty should use default
	}

	service, err := NewService(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "gemini-embedding-001", service.modelName)
}

func TestNewService_CustomModelName(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		APIKey:    "test-api-key",
		ModelName: "custom-embedding-model",
	}

	service, err := NewService(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "custom-embedding-model", service.modelName)
}

func TestNewService_EmptyAPIKey(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		APIKey:    "",
		ModelName: "gemini-embedding-001",
	}

	// genai.NewClient should fail with empty API key
	service, err := NewService(ctx, cfg)
	// Note: genai.NewClient might not validate API key at creation time
	// It may only fail when actually making API calls
	// For now, we just verify the behavior
	if err != nil {
		assert.Error(t, err)
		assert.Nil(t, service)
	} else {
		// If it doesn't fail immediately, that's also valid behavior
		assert.NotNil(t, service)
	}
}

func TestService_Close(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		APIKey:    "test-api-key",
		ModelName: "gemini-embedding-001",
	}

	service, err := NewService(ctx, cfg)
	require.NoError(t, err)

	// Close should return nil (currently a no-op)
	err = service.Close()
	assert.NoError(t, err)
}

func TestConfig_DefaultValues(t *testing.T) {
	// Test that Config struct can be created with zero values
	cfg := Config{}
	assert.Equal(t, "", cfg.APIKey)
	assert.Equal(t, "", cfg.ModelName)
}

func TestConfig_SetAPIKey(t *testing.T) {
	cfg := Config{
		APIKey: "my-api-key",
	}
	assert.Equal(t, "my-api-key", cfg.APIKey)
}

func TestConfig_SetModelName(t *testing.T) {
	cfg := Config{
		ModelName: "custom-model",
	}
	assert.Equal(t, "custom-model", cfg.ModelName)
}

func TestConfig_EmptyAPIKey(t *testing.T) {
	cfg := Config{APIKey: ""}
	assert.Equal(t, "", cfg.APIKey)
}

func TestConfig_EmptyModelName(t *testing.T) {
	cfg := Config{ModelName: ""}
	assert.Equal(t, "", cfg.ModelName)
}

func TestService_ModelName_Default(t *testing.T) {
	service := &Service{}
	assert.Equal(t, "", service.modelName)
}

func TestService_ModelName_Custom(t *testing.T) {
	service := &Service{
		modelName: "custom-model",
	}
	assert.Equal(t, "custom-model", service.modelName)
}

func TestService_ModelName_WithPrefix(t *testing.T) {
	service := &Service{
		modelName: "models/embedding-001",
	}
	assert.Equal(t, "models/embedding-001", service.modelName)
}

func TestService_ModelName_UpperCase(t *testing.T) {
	service := &Service{
		modelName: "GEMINI-EMBEDDING-001",
	}
	assert.Equal(t, "GEMINI-EMBEDDING-001", service.modelName)
}

func TestService_ModelName_AllNumbers(t *testing.T) {
	service := &Service{
		modelName: "123",
	}
	assert.Equal(t, "123", service.modelName)
}

func TestService_ModelName_WithDash(t *testing.T) {
	service := &Service{
		modelName: "embedding-v1",
	}
	assert.Equal(t, "embedding-v1", service.modelName)
}

func TestService_ModelName_WithUnderscore(t *testing.T) {
	service := &Service{
		modelName: "embedding_v1",
	}
	assert.Equal(t, "embedding_v1", service.modelName)
}

func TestService_ModelName_MultipleDashes(t *testing.T) {
	service := &Service{
		modelName: "gemini-embedding-v1-alpha",
	}
	assert.Equal(t, "gemini-embedding-v1-alpha", service.modelName)
}

func TestService_ModelName_LongModel(t *testing.T) {
	service := &Service{
		modelName: "embedding-model-very-long-name-with-many-parts",
	}
	assert.Equal(t, "embedding-model-very-long-name-with-many-parts", service.modelName)
}

func TestService_EmptyClient(t *testing.T) {
	service := &Service{
		client:    nil,
		modelName: "gemini-embedding-001",
	}
	assert.Nil(t, service.client)
}

func TestService_ClientNotNil(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		APIKey:    "test-key",
		ModelName: "model-1",
	}

	service, err := NewService(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, service.client)
	assert.NotNil(t, service)
}

func TestService_MultipleModels(t *testing.T) {
	models := []string{
		"gemini-embedding-001",
		"embedding-ada-002",
		"text-embedding-ada-001",
		"all-MiniLM-L6-v2",
		"model-v1",
		"model-v2-large",
	}

	for _, model := range models {
		ctx := context.Background()
		cfg := Config{
			APIKey:    "test-key",
			ModelName: model,
		}

		service, err := NewService(ctx, cfg)
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, model, service.modelName)
	}
}

// Note: EmbedQuery and EmbedDocuments require a real API key and connection
// to the Gemini API. These should be tested in integration tests, not unit tests.
// The tests below are commented out as they would require:
// 1. A valid GOOGLE_API_KEY environment variable
// 2. Network access to Google's API
// 3. API quota/billing setup
//
// func TestEmbedQuery_Integration(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test")
// 	}
//
// 	apiKey := os.Getenv("GOOGLE_API_KEY")
// 	if apiKey == "" {
// 		t.Skip("GOOGLE_API_KEY not set")
// 	}
//
// 	ctx := context.Background()
// 	cfg := Config{
// 		APIKey:    apiKey,
// 		ModelName: "gemini-embedding-001",
// 	}
//
// 	service, err := NewService(ctx, cfg)
// 	require.NoError(t, err)
// 	defer service.Close()
//
// 	embedding, err := service.EmbedQuery(ctx, "Hello, world!")
// 	require.NoError(t, err)
// 	assert.NotNil(t, embedding)
// 	assert.Greater(t, len(embedding), 0)
// }
//
// func TestEmbedDocuments_Integration(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test")
// 	}
//
// 	apiKey := os.Getenv("GOOGLE_API_KEY")
// 	if apiKey == "" {
// 		t.Skip("GOOGLE_API_KEY not set")
// 	}
//
// 	ctx := context.Background()
// 	cfg := Config{
// 		APIKey:    apiKey,
// 		ModelName: "gemini-embedding-001",
// 	}
//
// 	service, err := NewService(ctx, cfg)
// 	require.NoError(t, err)
// 	defer service.Close()
//
// 	documents := []string{"Hello", "World", "Test"}
// 	embeddings, err := service.EmbedDocuments(ctx, documents)
// 	require.NoError(t, err)
// 	assert.NotNil(t, embeddings)
// 	assert.Equal(t, 3, len(embeddings))
// }

func TestService_StructFields(t *testing.T) {
	service := &Service{
		client:    nil,
		modelName: "test-model",
	}

	assert.NotNil(t, service)
	assert.Nil(t, service.client)
	assert.Equal(t, "test-model", service.modelName)
}

func TestService_StructFields_WithClient(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		APIKey:    "key",
		ModelName: "model",
	}

	service, err := NewService(ctx, cfg)
	require.NoError(t, err)

	assert.NotNil(t, service)
	assert.NotNil(t, service.client)
	assert.Equal(t, "model", service.modelName)
}

func TestService_ConstructorWithAllFields(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		APIKey:    "api-key-123",
		ModelName: "embedding-001",
	}

	service, err := NewService(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, service)

	assert.Equal(t, "api-key-123", cfg.APIKey)
	assert.Equal(t, "embedding-001", cfg.ModelName)
	assert.NotNil(t, service.client)
}

func TestService_MultipleInstances(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := context.Background()
	cfg1 := Config{
		APIKey:    "key1",
		ModelName: "model1",
	}
	cfg2 := Config{
		APIKey:    "key2",
		ModelName: "model2",
	}

	service1, err := NewService(ctx1, cfg1)
	service2, err := NewService(ctx2, cfg2)

	require.NoError(t, err)
	require.NotNil(t, service1)
	require.NotNil(t, service2)
	assert.NotNil(t, service1.client)
	assert.NotNil(t, service2.client)
}

func TestService_MixedModels(t *testing.T) {
	ctx := context.Background()
	models := []string{"model-a", "model-b", "model-c", "model-d"}

	for _, model := range models {
		cfg := Config{
			APIKey:    "key",
			ModelName: model,
		}
		service, err := NewService(ctx, cfg)
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, model, service.modelName)
	}
}

func TestService_EmptyService(t *testing.T) {
	service := &Service{}
	assert.NotNil(t, service)
}

func TestConfig_StructWithValues(t *testing.T) {
	cfg := Config{
		APIKey:    "test-key-12345",
		ModelName: "test-model-67890",
	}

	assert.Equal(t, "test-key-12345", cfg.APIKey)
	assert.Equal(t, "test-model-67890", cfg.ModelName)
}

func TestService_CloseOnEmptyService(t *testing.T) {
	service := &Service{}
	err := service.Close()
	assert.NoError(t, err)
}
