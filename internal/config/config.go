// Package config handles application configuration using koanf.
package config

import (
	"log"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config holds all application configuration.
type Config struct {
	Model       ModelConfig       `koanf:"model"`
	Agent       AgentConfig       `koanf:"agent"`
	VectorStore VectorStoreConfig `koanf:"vectorstore"`
	Retriever   RetrieverConfig   `koanf:"retriever"`
	Server      ServerConfig      `koanf:"server"`
	Tracing     TracingConfig     `koanf:"tracing"`
}

// ModelConfig holds LLM model settings.
type ModelConfig struct {
	Name           string  `koanf:"name"`
	EmbeddingModel string  `koanf:"embedding_model"`
	APIKey         string  `koanf:"api_key"`
	Temperature    float64 `koanf:"temperature"`
	MaxTokens      int     `koanf:"max_tokens"`
}

// AgentConfig holds agent settings.
type AgentConfig struct {
	Name        string `koanf:"name"`
	Description string `koanf:"description"`
	Instruction string `koanf:"instruction"`
}

// VectorStoreConfig holds vector database settings.
type VectorStoreConfig struct {
	Provider   string `koanf:"provider"`
	URL        string `koanf:"url"`
	GRPCPort   int    `koanf:"grpc_port"`
	Collection string `koanf:"collection"`
	VectorSize uint64 `koanf:"vector_size"`
	APIKey     string `koanf:"api_key"`
}

// RetrieverConfig holds retrieval settings.
type RetrieverConfig struct {
	TopK         int     `koanf:"top_k"`
	MinScore     float64 `koanf:"min_score"`
	ChunkSize    int     `koanf:"chunk_size"`
	ChunkOverlap int     `koanf:"chunk_overlap"`
}

// ServerConfig holds server settings.
type ServerConfig struct {
	Host       string `koanf:"host"`
	Port       int    `koanf:"port"`
	APIKey     string `koanf:"api_key"`
	RateLimit  int    `koanf:"rate_limit"`
	RateWindow int    `koanf:"rate_window"`
}

// TracingConfig holds OpenTelemetry tracing settings.
type TracingConfig struct {
	Enabled     bool   `koanf:"enabled"`
	Endpoint    string `koanf:"endpoint"`
	ServiceName string `koanf:"service_name"`
}

// Load loads configuration from files and environment variables.
// Priority (highest to lowest): env vars > config.yaml > defaults
func Load(configPath string) (*Config, error) {
	// Create a fresh koanf instance for each load
	k := koanf.New(".")

	// Set defaults
	cfg := &Config{
		Model: ModelConfig{
			Name:           "gemini-2.5-flash",
			EmbeddingModel: "gemini-embedding-001",
			Temperature:    0.7,
			MaxTokens:      2048,
		},
		Agent: AgentConfig{
			Name:        "rag_agent",
			Description: "An intelligent RAG agent.",
			Instruction: "You are a helpful RAG assistant.",
		},
		VectorStore: VectorStoreConfig{
			Provider:   "qdrant",
			URL:        "localhost",
			GRPCPort:   6334,
			Collection: "agenticraggo",
			VectorSize: 768, // Default for many embedding models
		},
		Retriever: RetrieverConfig{
			TopK:         10,
			MinScore:     0.7,
			ChunkSize:    512,
			ChunkOverlap: 50,
		},
		Server: ServerConfig{
			Host:       "0.0.0.0",
			Port:       8001,
			APIKey:     "",
			RateLimit:  100,
			RateWindow: 60,
		},
		Tracing: TracingConfig{
			Enabled:     false,
			Endpoint:    "http://localhost:4317",
			ServiceName: "agentic-rag-go",
		},
	}

	// Load from YAML config file (if exists)
	if configPath != "" {
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			log.Printf("Warning: could not load config file %s: %v", configPath, err)
		}
	}

	// Load from environment variables (prefix: APP_)
	// e.g., APP_MODEL_NAME, APP_VECTORSTORE_URL
	if err := k.Load(env.Provider("APP_", ".", func(s string) string {
		return strings.Replace(
			strings.ToLower(strings.TrimPrefix(s, "APP_")),
			"_", ".", -1,
		)
	}), nil); err != nil {
		return nil, err
	}

	// Also check for common env vars without prefix
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" && cfg.Model.APIKey == "" {
		cfg.Model.APIKey = apiKey
	}
	if qdrantURL := os.Getenv("QDRANT_URL"); qdrantURL != "" {
		cfg.VectorStore.URL = qdrantURL
	}
	if phoenixEndpoint := os.Getenv("PHOENIX_COLLECTOR_ENDPOINT"); phoenixEndpoint != "" {
		cfg.Tracing.Endpoint = phoenixEndpoint
		cfg.Tracing.Enabled = true
	}
	if otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); otelEndpoint != "" {
		cfg.Tracing.Endpoint = otelEndpoint
		cfg.Tracing.Enabled = true
	}

	// Unmarshal into config struct
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// MustLoad loads configuration or panics on error.
func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}
