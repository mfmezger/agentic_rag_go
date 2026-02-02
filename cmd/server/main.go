package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mfmezger/agentic_rag_go/internal/api"
	"github.com/mfmezger/agentic_rag_go/internal/config"

	_ "github.com/mfmezger/agentic_rag_go/docs" // Import generated swagger docs

	"github.com/joho/godotenv"
)

//	@title			Agentic RAG API
//	@version		1.0
//	@description	REST API for Retrieval-Augmented Generation with Qdrant vector store
//	@host			localhost:8001
//	@BasePath		/api/v1

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Configuration loaded:")
	log.Printf("  Server: %s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("  VectorStore: %s:%d (collection: %s)", cfg.VectorStore.URL, cfg.VectorStore.GRPCPort, cfg.VectorStore.Collection)
	log.Printf("  Chunking: size=%d, overlap=%d", cfg.Retriever.ChunkSize, cfg.Retriever.ChunkOverlap)

	// Create API server
	server, err := api.NewServer(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
		os.Exit(0)
	}()

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}