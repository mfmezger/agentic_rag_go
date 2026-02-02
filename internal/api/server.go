// Package api provides the REST API for the RAG service.
//
//	@title			Agentic RAG API
//	@version		1.0
//	@description	REST API for Retrieval-Augmented Generation with Qdrant vector store
//	@host			localhost:8001
//	@BasePath		/api/v1
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	ragagent "github.com/mfmezger/agentic_rag_go/internal/agent"
	"github.com/mfmezger/agentic_rag_go/internal/config"
	"github.com/mfmezger/agentic_rag_go/internal/vectorstore/qdrant"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"

	"github.com/google/uuid"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/tmc/langchaingo/textsplitter"
)

// Server is the REST API server.
type Server struct {
	cfg          *config.Config
	qdrant       *qdrant.Client
	mux          *http.ServeMux
	splitter     textsplitter.TextSplitter
	agentFactory *ragagent.Factory
	middleware   *middleware
	apiVersion   string
}

// NewServer creates a new API server.
func NewServer(ctx context.Context, cfg *config.Config) (*Server, error) {
	// Initialize Qdrant client
	qdrantClient, err := qdrant.New(ctx, qdrant.Config{
		Host:       cfg.VectorStore.URL,
		GRPCPort:   cfg.VectorStore.GRPCPort,
		Collection: cfg.VectorStore.Collection,
		VectorSize: cfg.VectorStore.VectorSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	// Ensure collection exists
	if err := qdrantClient.EnsureCollection(ctx, cfg.VectorStore.Collection, cfg.VectorStore.VectorSize); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	// Create text splitter using langchaingo
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(cfg.Retriever.ChunkSize),
		textsplitter.WithChunkOverlap(cfg.Retriever.ChunkOverlap),
	)

	// Create agent factory
	agentFactory, err := ragagent.NewFactory(ctx, cfg, qdrantClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent factory: %w", err)
	}

	s := &Server{
		cfg:          cfg,
		qdrant:       qdrantClient,
		mux:          http.NewServeMux(),
		splitter:     splitter,
		agentFactory: agentFactory,
		middleware: newMiddleware(
			cfg.Server.APIKey,
			cfg.Server.RateLimit,
			time.Duration(cfg.Server.RateWindow)*time.Second,
		),
		apiVersion: "v1",
	}

	// Register routes
	s.registerRoutes()

	return s, nil
}

// registerRoutes sets up all API routes.
func (s *Server) registerRoutes() {
	v1Prefix := "/api/" + s.apiVersion

	s.mux.HandleFunc("GET /health", s.handleHealth)

	s.mux.HandleFunc("POST "+v1Prefix+"/upload_text",
		s.middleware.rateLimit(s.middleware.auth(s.handleUploadText)))
	s.mux.HandleFunc("POST "+v1Prefix+"/search",
		s.middleware.rateLimit(s.middleware.auth(s.handleSearch)))
	s.mux.HandleFunc("POST "+v1Prefix+"/chat",
		s.middleware.rateLimit(s.middleware.auth(s.handleChat)))

	s.mux.HandleFunc("POST "+v1Prefix+"/documents/upload",
		s.middleware.rateLimit(s.middleware.auth(s.handleUploadTextV2)))
	s.mux.HandleFunc("POST "+v1Prefix+"/documents/search",
		s.middleware.rateLimit(s.middleware.auth(s.handleSearchV2)))
	s.mux.HandleFunc("POST "+v1Prefix+"/conversations/chat",
		s.middleware.rateLimit(s.middleware.auth(s.handleChatV2)))

	s.mux.Handle("GET /docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

// Close cleans up server resources.
func (s *Server) Close() error {
	if s.qdrant != nil {
		return s.qdrant.Close()
	}
	return nil
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	log.Printf("Starting API server on %s", addr)
	log.Printf("Swagger docs available at http://%s/docs/", addr)
	return http.ListenAndServe(addr, s)
}

// handleHealth returns server health status.
//
//	@Summary		Health check
//	@Description	Returns the health status of the API
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// UploadTextRequest is the request body for upload_text.
type UploadTextRequest struct {
	Text     string            `json:"text" example:"Your document text goes here..."`
	Metadata map[string]string `json:"metadata,omitempty" example:"author:John Doe"`
	Source   string            `json:"source,omitempty" example:"document.pdf"`
}

// UploadTextResponse is the response for upload_text.
type UploadTextResponse struct {
	Message    string   `json:"message" example:"Text uploaded and chunked successfully"`
	ChunkCount int      `json:"chunk_count" example:"5"`
	ChunkIDs   []string `json:"chunk_ids"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request body"`
}

// handleUploadText handles the POST /api/v1/upload_text endpoint.
//
//	@Summary		Upload text
//	@Description	Chunks text and stores it in the vector database for retrieval
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Param			request	body		UploadTextRequest	true	"Text to upload"
//	@Success		200		{object}	UploadTextResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/upload_text [post]
func (s *Server) handleUploadText(w http.ResponseWriter, r *http.Request) {
	var req UploadTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.Text == "" {
		s.writeError(w, http.StatusBadRequest, "Text field is required")
		return
	}

	// Split text into chunks using langchaingo
	chunks, err := s.splitter.SplitText(req.Text)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to split text: "+err.Error())
		return
	}

	if len(chunks) == 0 {
		s.writeError(w, http.StatusBadRequest, "No chunks generated from text")
		return
	}

	// Generate embeddings for all chunks using Gemini
	embeddings, err := s.agentFactory.EmbeddingService().EmbedDocuments(r.Context(), chunks)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to generate embeddings: "+err.Error())
		return
	}

	// Prepare documents for Qdrant
	docs := make([]qdrant.Document, len(chunks))
	chunkIDs := make([]string, len(chunks))

	for i, chunk := range chunks {
		id := uuid.New().String()
		chunkIDs[i] = id

		// Merge metadata
		metadata := make(map[string]string)
		for k, v := range req.Metadata {
			metadata[k] = v
		}
		if req.Source != "" {
			metadata["source"] = req.Source
		}
		metadata["chunk_index"] = fmt.Sprintf("%d", i)

		docs[i] = qdrant.Document{
			ID:       id,
			Content:  chunk,
			Metadata: metadata,
			Dense:    embeddings[i],
			Sparse:   nil, // TODO: Add BM25 sparse vector for hybrid search
		}
	}

	// Store in Qdrant
	if err := s.qdrant.Upsert(r.Context(), s.cfg.VectorStore.Collection, docs); err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to store documents: "+err.Error())
		return
	}

	log.Printf("Uploaded %d chunks from text (source: %s)", len(chunks), req.Source)

	s.writeJSON(w, http.StatusOK, UploadTextResponse{
		Message:    "Text uploaded and chunked successfully",
		ChunkCount: len(chunks),
		ChunkIDs:   chunkIDs,
	})
}

// SearchRequest is the request body for search.
type SearchRequest struct {
	Query string `json:"query" example:"What is machine learning?"`
	TopK  int    `json:"top_k,omitempty" example:"5"`
}

// SearchResponse is the response for search.
type SearchResponse struct {
	Results []SearchResultItem `json:"results"`
}

// SearchResultItem is a single search result.
type SearchResultItem struct {
	ID       string            `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Content  string            `json:"content" example:"Machine learning is..."`
	Score    float32           `json:"score" example:"0.95"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// handleSearch handles the POST /api/v1/search endpoint.
//
//	@Summary		Search documents
//	@Description	Search for documents using hybrid vector search
//	@Tags			search
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SearchRequest	true	"Search query"
//	@Success		200		{object}	SearchResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/search [post]
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.Query == "" {
		s.writeError(w, http.StatusBadRequest, "Query field is required")
		return
	}

	topK := req.TopK
	if topK <= 0 {
		topK = s.cfg.Retriever.TopK
	}

	// Generate query embedding using Gemini
	queryVector, err := s.agentFactory.EmbeddingService().EmbedQuery(r.Context(), req.Query)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to generate query embedding: "+err.Error())
		return
	}

	results, err := s.qdrant.HybridSearch(r.Context(), s.cfg.VectorStore.Collection, queryVector, nil, uint64(topK))
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Search failed: "+err.Error())
		return
	}

	items := make([]SearchResultItem, len(results))
	for i, r := range results {
		items[i] = SearchResultItem{
			ID:       r.ID,
			Content:  r.Content,
			Score:    r.Score,
			Metadata: r.Payload,
		}
	}

	s.writeJSON(w, http.StatusOK, SearchResponse{Results: items})
}

// writeJSON writes a JSON response.
func (s *Server) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) handleUploadTextV2(w http.ResponseWriter, r *http.Request) {
	s.handleUploadText(w, r)
}

func (s *Server) handleSearchV2(w http.ResponseWriter, r *http.Request) {
	s.handleSearch(w, r)
}

func (s *Server) handleChatV2(w http.ResponseWriter, r *http.Request) {
	s.handleChat(w, r)
}

// ChatRequest is the request body for chat.
type ChatRequest struct {
	Message   string `json:"message" example:"What is machine learning?"`
	SessionID string `json:"session_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string `json:"user_id,omitempty" example:"user123"`
}

// ChatResponse is the response for chat.
type ChatResponse struct {
	Response  string `json:"response"`
	SessionID string `json:"session_id"`
}

// handleChat handles the POST /api/v1/chat endpoint.
//
//	@Summary		Chat with RAG agent
//	@Description	Send a message to the RAG agent which searches internal docs first, then web
//	@Tags			chat
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ChatRequest	true	"Chat message"
//	@Success		200		{object}	ChatResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/chat [post]
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.Message == "" {
		s.writeError(w, http.StatusBadRequest, "Message field is required")
		return
	}

	// Set defaults
	userID := req.UserID
	if userID == "" {
		userID = "default_user"
	}

	ctx := r.Context()

	// Create or get session
	sessionID := req.SessionID
	sessionService := s.agentFactory.SessionService()

	if sessionID == "" {
		// Create new session
		resp, err := sessionService.Create(ctx, &session.CreateRequest{
			AppName: "agentic_rag_go",
			UserID:  userID,
		})
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "Failed to create session: "+err.Error())
			return
		}
		sessionID = resp.Session.ID()
	}

	// Pre-fetch documents (cheap operation - runs before agent)
	retrieved, err := s.agentFactory.Retrieve(ctx, req.Message)
	if err != nil {
		log.Printf("Warning: retrieval failed: %v", err)
		// Continue without retrieved context - agent can still use GoogleSearch
	}

	// Create runner with pre-fetched context
	runner, err := s.agentFactory.NewRunner(ctx, "agentic_rag_go", retrieved)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, "Failed to create runner: "+err.Error())
		return
	}

	// Run agent
	userMsg := genai.NewContentFromText(req.Message, genai.RoleUser)
	var responseText string

	for event, err := range runner.Run(ctx, userID, sessionID, userMsg, agent.RunConfig{}) {
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, "Agent error: "+err.Error())
			return
		}
		if event.LLMResponse.Content == nil {
			continue
		}
		for _, p := range event.LLMResponse.Content.Parts {
			if p.Text != "" {
				responseText += p.Text
			}
		}
	}

	s.writeJSON(w, http.StatusOK, ChatResponse{
		Response:  responseText,
		SessionID: sessionID,
	})
}
