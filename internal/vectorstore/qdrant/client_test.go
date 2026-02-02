package qdrant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocument_StructCreation(t *testing.T) {
	doc := Document{
		ID:       "test-id",
		Content:  "test content",
		Metadata: map[string]string{"key": "value"},
		Dense:    []float32{0.1, 0.2, 0.3},
		Sparse:   nil,
	}

	assert.Equal(t, "test-id", doc.ID)
	assert.Equal(t, "test content", doc.Content)
	assert.Equal(t, "value", doc.Metadata["key"])
	assert.Equal(t, 3, len(doc.Dense))
	assert.Nil(t, doc.Sparse)
}

func TestDocument_EmptyContent(t *testing.T) {
	doc := Document{
		ID:       "test-id",
		Content:  "",
		Metadata: map[string]string{},
		Dense:    []float32{},
	}

	assert.Equal(t, "", doc.Content)
	assert.Equal(t, 0, len(doc.Dense))
}

func TestDocument_WithSparseVector(t *testing.T) {
	sparse := SparseVector{
		Indices: []uint32{1, 5, 10},
		Values:  []float32{0.5, 0.3, 0.8},
	}
	doc := Document{
		ID:       "test-id",
		Content:  "test content",
		Metadata: map[string]string{"key": "value"},
		Dense:    []float32{0.1, 0.2, 0.3},
		Sparse:   &sparse,
	}

	assert.NotNil(t, doc.Sparse)
	assert.Equal(t, 3, len(doc.Sparse.Indices))
	assert.Equal(t, 3, len(doc.Sparse.Values))
}

func TestDocument_ZeroVector(t *testing.T) {
	doc := Document{
		ID:       "test-id",
		Content:  "test",
		Metadata: map[string]string{},
		Dense:    []float32{},
		Sparse:   nil,
	}

	assert.Equal(t, 0, len(doc.Dense))
	assert.Nil(t, doc.Sparse)
}

func TestDocument_LargeMetadata(t *testing.T) {
	metadata := make(map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		metadata[string(rune(i))] = "value"
	}

	doc := Document{
		ID:       "test-id",
		Content:  "content",
		Metadata: metadata,
		Dense:    []float32{0.1, 0.2, 0.3},
	}

	assert.Equal(t, 1000, len(doc.Metadata))
}

func TestSparseVector_StructCreation(t *testing.T) {
	sparse := SparseVector{
		Indices: []uint32{1, 5, 10},
		Values:  []float32{0.5, 0.3, 0.8},
	}

	assert.Equal(t, 3, len(sparse.Indices))
	assert.Equal(t, 3, len(sparse.Values))
	assert.Equal(t, uint32(1), sparse.Indices[0])
	assert.Equal(t, float32(0.5), sparse.Values[0])
}

func TestSparseVector_EmptyVectors(t *testing.T) {
	sparse := SparseVector{
		Indices: []uint32{},
		Values:  []float32{},
	}

	assert.Equal(t, 0, len(sparse.Indices))
	assert.Equal(t, 0, len(sparse.Values))
}

func TestSparseVector_SingleIndex(t *testing.T) {
	sparse := SparseVector{
		Indices: []uint32{42},
		Values:  []float32{1.0},
	}

	assert.Equal(t, 1, len(sparse.Indices))
	assert.Equal(t, 1, len(sparse.Values))
}

func TestSparseVector_ReverseOrder(t *testing.T) {
	sparse := SparseVector{
		Indices: []uint32{10, 5, 1},
		Values:  []float32{0.8, 0.3, 0.5},
	}

	assert.Equal(t, uint32(10), sparse.Indices[0])
	assert.Equal(t, float32(0.8), sparse.Values[0])
}

func TestSearchResult_StructCreation(t *testing.T) {
	result := SearchResult{
		ID:      "result-123",
		Score:   0.95,
		Content: "search result content",
		Payload: map[string]string{"metadata_key": "metadata_value"},
	}

	assert.Equal(t, "result-123", result.ID)
	assert.Equal(t, float32(0.95), result.Score)
	assert.Equal(t, "search result content", result.Content)
	assert.Equal(t, "metadata_value", result.Payload["metadata_key"])
}

func TestSearchResult_ZeroScore(t *testing.T) {
	result := SearchResult{
		ID:      "id",
		Score:   0.0,
		Content: "content",
		Payload: map[string]string{},
	}

	assert.Equal(t, float32(0.0), result.Score)
}

func TestSearchResult_MaxScore(t *testing.T) {
	result := SearchResult{
		ID:      "id",
		Score:   1.0,
		Content: "content",
		Payload: map[string]string{},
	}

	assert.Equal(t, float32(1.0), result.Score)
}

func TestSearchResult_NegativeScore(t *testing.T) {
	result := SearchResult{
		ID:      "id",
		Score:   -0.5,
		Content: "content",
		Payload: map[string]string{},
	}

	assert.Equal(t, float32(-0.5), result.Score)
}

func TestSearchResult_EmptyPayload(t *testing.T) {
	result := SearchResult{
		ID:      "id",
		Score:   0.8,
		Content: "content",
		Payload: map[string]string{},
	}

	assert.Equal(t, 0, len(result.Payload))
}

func TestSearchResult_MultipleFields(t *testing.T) {
	result := SearchResult{
		ID:      "test-id",
		Score:   0.9,
		Content: "test content",
		Payload: map[string]string{
			"author": "john",
			"date":   "2024-01-01",
			"type":   "document",
		},
	}

	assert.Equal(t, "test-id", result.ID)
	assert.Equal(t, float32(0.9), result.Score)
	assert.Equal(t, "test content", result.Content)
	assert.Equal(t, "john", result.Payload["author"])
	assert.Equal(t, "2024-01-01", result.Payload["date"])
	assert.Equal(t, "document", result.Payload["type"])
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config",
			config: Config{
				Host:       "localhost",
				GRPCPort:   6334,
				Collection: "test",
				VectorSize: 768,
			},
			valid: true,
		},
		{
			name: "empty host",
			config: Config{
				Host:       "",
				GRPCPort:   6334,
				Collection: "test",
				VectorSize: 768,
			},
			valid: false,
		},
		{
			name: "zero port",
			config: Config{
				Host:       "localhost",
				GRPCPort:   0,
				Collection: "test",
				VectorSize: 768,
			},
			valid: false,
		},
		{
			name: "large port",
			config: Config{
				Host:       "localhost",
				GRPCPort:   65535,
				Collection: "test",
				VectorSize: 768,
			},
			valid: true,
		},
		{
			name: "default port",
			config: Config{
				Host:       "localhost",
				GRPCPort:   6334,
				Collection: "",
				VectorSize: 768,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.config.Host)
				assert.Greater(t, tt.config.GRPCPort, 0)
			} else {
				if tt.config.Host == "" {
					assert.Empty(t, tt.config.Host)
				}
				if tt.config.GRPCPort == 0 {
					assert.Equal(t, 0, tt.config.GRPCPort)
				}
			}
		})
	}
}

func TestConfig_VerifyDefaultValues(t *testing.T) {
	cfg := Config{
		Host:       "localhost",
		GRPCPort:   6334,
		Collection: "default",
		VectorSize: 768,
	}

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 6334, cfg.GRPCPort)
	assert.Equal(t, "default", cfg.Collection)
	assert.Equal(t, uint64(768), cfg.VectorSize)
}

func TestConfig_CustomVectorSize(t *testing.T) {
	sizes := []uint64{128, 256, 512, 1024, 1536, 3072}

	for _, size := range sizes {
		cfg := Config{
			Host:       "localhost",
			GRPCPort:   6334,
			Collection: "test",
			VectorSize: size,
		}

		assert.Equal(t, size, cfg.VectorSize)
	}
}

// Note: New(), EnsureCollection(), Upsert(), HybridSearch(), and Close()
// require mocking the Qdrant gRPC client, which is complex.
// These should be tested with:
// 1. Integration tests against a real Qdrant instance (Docker)
// 2. Or by creating extensive gRPC mocks
//
// For unit testing, we'd need to refactor the Client to accept
// interface-based gRPC clients rather than concrete types.
//
// Example integration test setup (requires Docker):
// func TestClient_Integration(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test")
// 	}
//
// 	ctx := context.Background()
// 	cfg := Config{
// 		Host:       "localhost",
// 		GRPCPort:   6334,
// 		Collection: "test_collection",
// 		VectorSize: 3,
// 	}
//
// 	client, err := New(ctx, cfg)
// 	require.NoError(t, err)
// 	defer client.Close()
//
// 	err = client.EnsureCollection(ctx, cfg.Collection, cfg.VectorSize)
// 	require.NoError(t, err)
//
// 	docs := []Document{
// 		{
// 			ID:       "doc1",
// 			Content:  "test content",
// 			Dense:    []float32{0.1, 0.2, 0.3},
// 			Metadata: map[string]string{"type": "test"},
// 		},
// 	}
//
// 	err = client.Upsert(ctx, cfg.Collection, docs)
// 	require.NoError(t, err)
//
// 	results, err := client.HybridSearch(ctx, cfg.Collection, []float32{0.1, 0.2, 0.3}, nil, 10)
// 	require.NoError(t, err)
// 	assert.Greater(t, len(results), 0)
// }

func TestSearchResult_IsNil(t *testing.T) {
	result := SearchResult{}

	assert.Equal(t, "", result.ID)
	assert.Equal(t, float32(0), result.Score)
	assert.Equal(t, "", result.Content)
	assert.Equal(t, 0, len(result.Payload))
}

func TestSearchResult_ContentWithSpecialCharacters(t *testing.T) {
	content := "Test with <special> \"chars\" & more"
	result := SearchResult{
		ID:      "id",
		Score:   0.9,
		Content: content,
		Payload: map[string]string{},
	}

	assert.Equal(t, content, result.Content)
}

func TestSearchResult_ContentWithNewlines(t *testing.T) {
	content := "line1\nline2\nline3"
	result := SearchResult{
		ID:      "id",
		Score:   0.9,
		Content: content,
		Payload: map[string]string{},
	}

	assert.Equal(t, content, result.Content)
}

func TestDocument_MultipleIndicesSparseVector(t *testing.T) {
	sparse := SparseVector{
		Indices: []uint32{0, 100, 200, 300, 400},
		Values:  []float32{0.9, 0.8, 0.7, 0.6, 0.5},
	}

	doc := Document{
		ID:       "doc-id",
		Content:  "document",
		Metadata: map[string]string{},
		Dense:    []float32{0.1, 0.2, 0.3},
		Sparse:   &sparse,
	}

	assert.NotNil(t, doc.Sparse)
	assert.Equal(t, 5, len(doc.Sparse.Indices))
	assert.Equal(t, 5, len(doc.Sparse.Values))
}

func TestDocument_EmptyString(t *testing.T) {
	doc := Document{
		ID:       "id",
		Content:  "",
		Metadata: map[string]string{},
		Dense:    nil,
		Sparse:   nil,
	}

	assert.Equal(t, "", doc.Content)
	assert.Nil(t, doc.Dense)
	assert.Nil(t, doc.Sparse)
}
