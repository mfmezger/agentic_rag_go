// Package agent provides RAG agent creation and configuration.
package agent

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mfmezger/agentic_rag_go/internal/config"
	"github.com/mfmezger/agentic_rag_go/internal/embedding"
	"github.com/mfmezger/agentic_rag_go/internal/vectorstore/qdrant"

	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"
	"google.golang.org/genai"
)

// Factory creates RAG agent runners.
type Factory struct {
	cfg            *config.Config
	qdrant         *qdrant.Client
	embedding      *embedding.Service
	model          model.LLM
	sessionService session.Service
}

// NewFactory creates a new agent factory.
func NewFactory(ctx context.Context, cfg *config.Config, qdrantClient *qdrant.Client) (*Factory, error) {
	// Initialize API key
	apiKey := cfg.Model.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}

	// Initialize Embedding Service
	embeddingService, err := embedding.NewService(ctx, embedding.Config{
		APIKey:    apiKey,
		ModelName: cfg.Model.EmbeddingModel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding service: %w", err)
	}

	// Initialize LLM Model
	llmModel, err := gemini.NewModel(ctx, cfg.Model.Name, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	return &Factory{
		cfg:            cfg,
		qdrant:         qdrantClient,
		embedding:      embeddingService,
		model:          llmModel,
		sessionService: session.InMemoryService(),
	}, nil
}

// RetrievedContext holds the pre-fetched documents for a query.
type RetrievedContext struct {
	Documents []qdrant.SearchResult
	Query     string
}

// Retrieve performs upfront document retrieval for a query.
// This should be called before NewRunner to pre-fetch relevant context.
func (f *Factory) Retrieve(ctx context.Context, query string) (*RetrievedContext, error) {
	topK := f.cfg.Retriever.TopK
	if topK <= 0 {
		topK = 10
	}

	// Generate query embedding using Gemini
	queryVector, err := f.embedding.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embedding query failed: %w", err)
	}

	results, err := f.qdrant.HybridSearch(ctx, f.cfg.VectorStore.Collection, queryVector, nil, uint64(topK))
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	return &RetrievedContext{
		Documents: results,
		Query:     query,
	}, nil
}

// NewRunner creates a new runner for the RAG agent.
// The retrieved context is injected into the agent's instruction.
// The agent only has GoogleSearch for web fallback (no function tool mixing).
func (f *Factory) NewRunner(ctx context.Context, appName string, retrieved *RetrievedContext) (*runner.Runner, error) {
	// Build context from retrieved documents
	var contextBuilder strings.Builder
	if retrieved != nil && len(retrieved.Documents) > 0 {
		contextBuilder.WriteString("\n\n## Retrieved Knowledge Base Documents\n\n")
		for i, doc := range retrieved.Documents {
			contextBuilder.WriteString(fmt.Sprintf("### Document %d (Score: %.2f)\n", i+1, doc.Score))
			contextBuilder.WriteString(doc.Content)
			contextBuilder.WriteString("\n\n")
		}
	}

	// Build instruction with injected context
	instruction := fmt.Sprintf(`%s

%s

STRATEGY:
1. First, analyze the retrieved documents above to answer the user's question.
2. If the retrieved documents contain sufficient information, use them to formulate your answer.
3. If the retrieved documents are insufficient or the topic requires current/real-time information, use google_search.
4. Always indicate whether your answer comes from internal documents or web search.
5. Cite sources when possible.
`, f.cfg.Agent.Instruction, contextBuilder.String())

	// Create agent with only GoogleSearch (native Gemini tool)
	// This avoids the function tool + native tool mixing issue
	ragAgent, err := llmagent.New(llmagent.Config{
		Name:        f.cfg.Agent.Name,
		Model:       f.model,
		Description: f.cfg.Agent.Description,
		Instruction: instruction,
		Tools: []tool.Tool{
			geminitool.GoogleSearch{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Create Runner
	r, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          ragAgent,
		SessionService: f.sessionService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return r, nil
}

// EmbeddingService returns the embedding service for use by other components.
func (f *Factory) EmbeddingService() *embedding.Service {
	return f.embedding
}

// SessionService returns the session service.
func (f *Factory) SessionService() session.Service {
	return f.sessionService
}
