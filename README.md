# Agentic RAG Go

An intelligent RAG (Retrieval-Augmented Generation) agent built with Google's Agent Development Kit (ADK) for Go.

## Project Structure

```bash
agentic_rag_go/
├── cmd/
│   └── server/           # Application entrypoint
│       └── main.go
├── internal/             # Private application code
│   ├── agent/            # Agent definitions and configuration
│   ├── config/           # Configuration loading
│   ├── retriever/        # RAG retrieval logic
│   ├── tools/            # Custom tools for agents
│   └── vectorstore/      # Vector database integrations
│       └── qdrant/       # Qdrant implementation
├── pkg/                  # Public libraries (if needed)
├── api/                  # API definitions
├── configs/              # Configuration files
├── scripts/              # Build/deployment scripts
├── docs/                 # Documentation
├── Makefile              # Build automation
└── go.mod                # Go module definition
```

## Quick Start

1. **Setup**

   ```bash
   chmod +x scripts/setup.sh
   ./scripts/setup.sh
   ```

2. **Configure**

   ```bash
   export GOOGLE_API_KEY="your-api-key"
   ```

3. **Run**

   ```bash
   make run
   ```

## Development

```bash
# Build the binary
make build

# Run tests
make test

# Lint code
make lint

# Format code
make fmt

# Show all commands
make help
```

## Configuration

See [configs/config.example.yaml](configs/config.example.yaml) for all available configuration options.

## License

MIT