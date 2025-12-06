# dago-adapters

[![Go Reference](https://pkg.go.dev/badge/github.com/aescanero/dago-adapters.svg)](https://pkg.go.dev/github.com/aescanero/dago-adapters)
[![Go Report Card](https://goreportcard.com/badge/github.com/aescanero/dago-adapters)](https://goreportcard.com/report/github.com/aescanero/dago-adapters)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**Shared adapter implementations for DA Orchestrator**

## Overview

`dago-adapters` provides concrete implementations (adapters) of the port interfaces defined in [`dago-libs`](https://github.com/aescanero/dago-libs). This repository contains all infrastructure adapters that are shared across DA Orchestrator components.

**Architecture Pattern**: Hexagonal Architecture (Ports & Adapters)
- **Ports** (interfaces) → [`dago-libs/pkg/ports/`](https://github.com/aescanero/dago-libs/tree/main/pkg/ports)
- **Adapters** (implementations) → This repository

## Available Adapters

### LLM Providers
- **Anthropic** - Claude models (Sonnet, Opus, Haiku)
- **OpenAI** - GPT models (GPT-4, GPT-4o, etc.)
- **Gemini** - Google's Gemini models
- **Ollama** - Local LLM execution

### Event Bus
- **Redis Streams** - Production-ready event bus
- **Memory** - In-memory event bus for testing

### Storage
- **Redis** - State persistence with Redis
- **Memory** - In-memory storage for testing

### Metrics
- **Prometheus** - Metrics collection and exposition

## Installation

```bash
go get github.com/aescanero/dago-adapters@latest
```

## Quick Start

### Using LLM Adapters

```go
import (
    "github.com/aescanero/dago-libs/pkg/ports"
    "github.com/aescanero/dago-adapters/pkg/llm"
)

// Create an LLM client using the factory
client, err := llm.NewClient(&llm.Config{
    Provider: "anthropic",  // or "openai", "gemini", "ollama"
    APIKey:   "your-api-key",
    Logger:   logger,
})

// Use the client (implements ports.LLMClient)
resp, err := client.Complete(ctx, ports.CompletionRequest{
    Messages: []ports.Message{
        {Role: "user", Content: "Hello!"},
    },
    Model: "claude-sonnet-4-20250514",
})
```

### Using Event Bus Adapters

```go
import (
    "github.com/aescanero/dago-adapters/pkg/events/redis"
)

eventBus, err := redis.NewEventBus(&redis.Config{
    Addr: "localhost:6379",
})
```

### Using Storage Adapters

```go
import (
    "github.com/aescanero/dago-adapters/pkg/storage/redis"
)

storage, err := redis.NewStorage(&redis.Config{
    Addr: "localhost:6379",
})
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│              dago-libs (PORTS)                  │
│  Interfaces: LLMClient, EventBus, Storage, etc. │
└─────────────────────────────────────────────────┘
                     ▲
                     │ implements
                     │
┌─────────────────────────────────────────────────┐
│         dago-adapters (ADAPTERS)                │
│  Concrete implementations with external SDKs    │
└─────────────────────────────────────────────────┘
         ▲                  ▲                ▲
         │                  │                │
    ┌────┴────┐    ┌────────┴──────┐   ┌────┴────┐
    │  dago   │    │ dago-node-    │   │ dago-   │
    │ (core)  │    │ executor      │   │ node-   │
    │         │    │               │   │ router  │
    └─────────┘    └───────────────┘   └─────────┘
```

## Documentation

- **[Architecture Guide](docs/architecture.md)** - Adapter architecture and design patterns
- **[LLM Providers](docs/llm-providers.md)** - Detailed guide for each LLM provider
- **[Event Bus](docs/event-bus.md)** - Event bus implementations
- **[Storage](docs/storage.md)** - Storage adapter documentation
- **[Contributing](docs/contributing.md)** - How to add new adapters

## Development

```bash
# Clone the repository
git clone https://github.com/aescanero/dago-adapters.git
cd dago-adapters

# Install dependencies
make deps

# Run tests
make test

# Run linter
make lint

# Build
make build
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific adapter
go test ./pkg/llm/anthropic/...
```

## Environment Variables

### LLM Providers
```bash
# Anthropic
ANTHROPIC_API_KEY=sk-ant-xxx

# OpenAI
OPENAI_API_KEY=sk-xxx

# Gemini
GEMINI_API_KEY=xxx

# Ollama (local)
OLLAMA_BASE_URL=http://localhost:11434
```

### Event Bus (Redis)
```bash
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Versioning

This repository follows [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR**: Breaking changes in adapter interfaces or behavior
- **MINOR**: New adapters or backward-compatible features
- **PATCH**: Bug fixes and improvements

## Dependencies

External SDKs used by adapters:
- **Anthropic**: `github.com/anthropics/anthropic-sdk-go`
- **OpenAI**: `github.com/sashabaranov/go-openai`
- **Gemini**: `github.com/google/generative-ai-go`
- **Ollama**: `github.com/jmorganca/ollama-go`
- **Redis**: `github.com/redis/go-redis/v9`
- **Prometheus**: `github.com/prometheus/client_golang`

## Related Repositories

- [`dago-libs`](https://github.com/aescanero/dago-libs) - Shared libraries and port interfaces
- [`dago`](https://github.com/aescanero/dago) - Core orchestrator
- [`dago-node-executor`](https://github.com/aescanero/dago-node-executor) - Executor worker
- [`dago-node-router`](https://github.com/aescanero/dago-node-router) - Router worker

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please read our [Contributing Guide](docs/contributing.md) before submitting pull requests.

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/aescanero/dago-adapters/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aescanero/dago-adapters/discussions)

## Project

- **Domain**: disasterproject.com
- **Organization**: aescanero
- **Repository**: [github.com/aescanero/dago-adapters](https://github.com/aescanero/dago-adapters)
