// Package embedding provides text embedding functionality using Gemini.
package embedding

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// Service handles text embedding operations.
type Service struct {
	client    *genai.Client
	modelName string
}

// Config holds embedding service configuration.
type Config struct {
	APIKey    string
	ModelName string // e.g., "gemini-embedding-001"
}

// NewService creates a new embedding service.
func NewService(ctx context.Context, cfg Config) (*Service, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	modelName := cfg.ModelName
	if modelName == "" {
		modelName = "gemini-embedding-001"
	}

	return &Service{
		client:    client,
		modelName: modelName,
	}, nil
}

// EmbedQuery generates an embedding for a query string.
func (s *Service) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(query, genai.RoleUser),
	}

	result, err := s.client.Models.EmbedContent(ctx, s.modelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("embed content failed: %w", err)
	}

	if result.Embeddings == nil || len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return result.Embeddings[0].Values, nil
}

// EmbedDocuments generates embeddings for multiple documents.
func (s *Service) EmbedDocuments(ctx context.Context, documents []string) ([][]float32, error) {
	contents := make([]*genai.Content, len(documents))
	for i, doc := range documents {
		contents[i] = genai.NewContentFromText(doc, genai.RoleUser)
	}

	result, err := s.client.Models.EmbedContent(ctx, s.modelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("embed content failed: %w", err)
	}

	if result.Embeddings == nil || len(result.Embeddings) != len(documents) {
		return nil, fmt.Errorf("unexpected number of embeddings: got %d, expected %d",
			len(result.Embeddings), len(documents))
	}

	embeddings := make([][]float32, len(result.Embeddings))
	for i, emb := range result.Embeddings {
		embeddings[i] = emb.Values
	}

	return embeddings, nil
}

// Close cleans up the embedding service resources.
func (s *Service) Close() error {
	// genai.Client doesn't have a Close method currently
	return nil
}
