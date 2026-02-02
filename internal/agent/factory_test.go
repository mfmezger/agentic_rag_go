package agent

import (
	"testing"

	"github.com/mfmezger/agentic_rag_go/internal/vectorstore/qdrant"
	"github.com/stretchr/testify/assert"
)

func TestRetrievedContext_StructCreation(t *testing.T) {
	docs := []qdrant.SearchResult{
		{
			ID:      "doc1",
			Score:   0.95,
			Content: "test content 1",
		},
		{
			ID:      "doc2",
			Score:   0.85,
			Content: "test content 2",
		},
	}

	retrieved := RetrievedContext{
		Documents: docs,
		Query:     "test query",
	}

	assert.Equal(t, 2, len(retrieved.Documents))
	assert.Equal(t, "test query", retrieved.Query)
	assert.Equal(t, float32(0.95), retrieved.Documents[0].Score)
}

func TestRetrievedContext_EmptyDocuments(t *testing.T) {
	retrieved := RetrievedContext{
		Documents: []qdrant.SearchResult{},
		Query:     "query with no results",
	}

	assert.Equal(t, 0, len(retrieved.Documents))
	assert.Equal(t, "query with no results", retrieved.Query)
}

func TestRetrievedContext_NilDocuments(t *testing.T) {
	retrieved := RetrievedContext{
		Documents: nil,
		Query:     "query",
	}

	assert.Nil(t, retrieved.Documents)
	assert.Equal(t, "query", retrieved.Query)
}

func TestRetrievedContext_DocumentOrdering(t *testing.T) {
	docs := []qdrant.SearchResult{
		{
			ID:      "doc1",
			Score:   0.70,
			Content: "low score",
		},
		{
			ID:      "doc2",
			Score:   0.95,
			Content: "high score",
		},
		{
			ID:      "doc3",
			Score:   0.80,
			Content: "medium score",
		},
	}

	retrieved := RetrievedContext{
		Documents: docs,
		Query:     "test",
	}

	assert.Equal(t, 3, len(retrieved.Documents))
	assert.Equal(t, "doc1", retrieved.Documents[0].ID)
	assert.Equal(t, "doc2", retrieved.Documents[1].ID)
	assert.Equal(t, "doc3", retrieved.Documents[2].ID)
}

func TestRetrievedContext_SingleDocument(t *testing.T) {
	retrieved := RetrievedContext{
		Documents: []qdrant.SearchResult{
			{
				ID:      "single-doc",
				Score:   1.0,
				Content: "only document",
			},
		},
		Query: "single query",
	}

	assert.Equal(t, 1, len(retrieved.Documents))
	assert.Equal(t, "single-doc", retrieved.Documents[0].ID)
	assert.Equal(t, float32(1.0), retrieved.Documents[0].Score)
}

func TestRetrievedContext_ZeroScores(t *testing.T) {
	retrieved := RetrievedContext{
		Documents: []qdrant.SearchResult{
			{
				ID:      "doc1",
				Score:   0.0,
				Content: "zero score",
			},
		},
		Query: "test",
	}

	assert.Equal(t, 1, len(retrieved.Documents))
	assert.Equal(t, float32(0.0), retrieved.Documents[0].Score)
}

func TestRetrievedContext_LargeQuery(t *testing.T) {
	longQuery := "this is a very long query that tests the handling of large query strings in the RetrievedContext struct"
	retrieved := RetrievedContext{
		Documents: []qdrant.SearchResult{},
		Query:     longQuery,
	}

	assert.Equal(t, longQuery, retrieved.Query)
	assert.Equal(t, 0, len(retrieved.Documents))
}

func TestRetrievedContext_EmptyQuery(t *testing.T) {
	retrieved := RetrievedContext{
		Documents: []qdrant.SearchResult{},
		Query:     "",
	}

	assert.Equal(t, "", retrieved.Query)
	assert.Equal(t, 0, len(retrieved.Documents))
}

func TestRetrievedContext_MixedScores(t *testing.T) {
	retrieved := RetrievedContext{
		Documents: []qdrant.SearchResult{
			{ID: "1", Score: 0.1, Content: "low"},
			{ID: "2", Score: 0.9, Content: "high"},
			{ID: "3", Score: 0.5, Content: "mid"},
			{ID: "4", Score: 0.99, Content: "highest"},
		},
		Query: "test",
	}

	assert.Equal(t, 4, len(retrieved.Documents))
	scores := []float32{0.1, 0.9, 0.5, 0.99}
	for i, doc := range retrieved.Documents {
		assert.Equal(t, scores[i], doc.Score)
	}
}

func TestRetrievedContext_DuplicateIDs(t *testing.T) {
	retrieved := RetrievedContext{
		Documents: []qdrant.SearchResult{
			{ID: "dup", Score: 0.8, Content: "first"},
			{ID: "dup", Score: 0.6, Content: "second"},
		},
		Query: "test",
	}

	assert.Equal(t, 2, len(retrieved.Documents))
	assert.Equal(t, "dup", retrieved.Documents[0].ID)
	assert.Equal(t, "dup", retrieved.Documents[1].ID)
}

func TestRetrievedContext_QuerySpecialCharacters(t *testing.T) {
	specialQuery := "What's the meaning of life? #hashtag @mention"
	retrieved := RetrievedContext{
		Documents: []qdrant.SearchResult{},
		Query:     specialQuery,
	}

	assert.Equal(t, specialQuery, retrieved.Query)
}

func TestRetrievedContext_ContentWithNewlines(t *testing.T) {
	retrieved := RetrievedContext{
		Documents: []qdrant.SearchResult{
			{
				ID:      "doc1",
				Score:   0.95,
				Content: "line1\nline2\nline3",
			},
		},
		Query: "test",
	}

	assert.Equal(t, "line1\nline2\nline3", retrieved.Documents[0].Content)
}

// Note: NewFactory(), Retrieve(), and NewRunner() require:
// 1. Valid API keys for Google Gemini
// 2. Running Qdrant instance
// 3. Complex mocking of ADK components (llmagent, model, session, runner)
//
// These are better suited for integration tests.
// To make them unit testable, we would need to:
// 1. Create interfaces for embedding service, qdrant client, LLM model
// 2. Accept these as constructor parameters
// 3. Inject mocks in tests
