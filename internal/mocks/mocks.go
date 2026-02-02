// Package mocks provides mock implementations for testing.
package mocks

import (
	"context"

	"github.com/mfmezger/agentic_rag_go/internal/embedding"
	"github.com/mfmezger/agentic_rag_go/internal/vectorstore/qdrant"
	"github.com/stretchr/testify/mock"
)

// MockEmbeddingService is a mock implementation of embedding.Service.
type MockEmbeddingService struct {
	mock.Mock
}

// EmbedQuery mocks the EmbedQuery method.
func (m *MockEmbeddingService) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]float32), args.Error(1)
}

// EmbedDocuments mocks the EmbedDocuments method.
func (m *MockEmbeddingService) EmbedDocuments(ctx context.Context, docs []string) ([][]float32, error) {
	args := m.Called(ctx, docs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]float32), args.Error(1)
}

// Close mocks the Close method.
func (m *MockEmbeddingService) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockQdrantClient is a mock implementation of qdrant.Client.
type MockQdrantClient struct {
	mock.Mock
}

// EnsureCollection mocks the EnsureCollection method.
func (m *MockQdrantClient) EnsureCollection(ctx context.Context, name string, vectorSize uint64) error {
	args := m.Called(ctx, name, vectorSize)
	return args.Error(0)
}

// Upsert mocks the Upsert method.
func (m *MockQdrantClient) Upsert(ctx context.Context, collection string, docs []qdrant.Document) error {
	args := m.Called(ctx, collection, docs)
	return args.Error(0)
}

// HybridSearch mocks the HybridSearch method.
func (m *MockQdrantClient) HybridSearch(ctx context.Context, collection string, denseVector []float32, sparseVector *qdrant.SparseVector, topK uint64) ([]qdrant.SearchResult, error) {
	args := m.Called(ctx, collection, denseVector, sparseVector, topK)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]qdrant.SearchResult), args.Error(1)
}

// Close mocks the Close method.
func (m *MockQdrantClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockTextSplitter is a mock implementation of textsplitter.TextSplitter.
type MockTextSplitter struct {
	mock.Mock
}

// SplitText mocks the SplitText method.
func (m *MockTextSplitter) SplitText(text string) ([]string, error) {
	args := m.Called(text)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockAgentFactory is a minimal mock for agent factory (simplified for testability).
type MockAgentFactory struct {
	mock.Mock
}

// EmbeddingService mocks the EmbeddingService method.
func (m *MockAgentFactory) EmbeddingService() *embedding.Service {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*embedding.Service)
}
