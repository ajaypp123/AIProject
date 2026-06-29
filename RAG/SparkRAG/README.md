# Spark RAG Platform

**Enterprise-Grade Spark Knowledge Assistant** powered by Retrieval Augmented Generation (RAG)

A production-ready RAG platform for querying Apache Spark, Hadoop, Iceberg, Delta Lake, and related big data technologies using local LLMs and embeddings.

## 📋 Table of Contents

- [Architecture](#architecture)
- [Features](#features)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Docker Deployment](#docker-deployment)
- [CLI Usage](#cli-usage)
- [API Reference](#api-reference)
- [Development](#development)
- [Troubleshooting](#troubleshooting)

## 🏗️ Architecture

```
┌─────────────────────┐
│  React Chat UI      │
│   (Port 3000)       │
└──────────┬──────────┘
           │ HTTP
           ▼
┌─────────────────────┐
│ Go REST API         │ ◄─────┐
│ (Port 8080)         │       │
│ • Chat              │       │
│ • Search            │       │
│ • Citations         │       │
└──────────┬──────────┘       │
           │                   │
    ┌──────┴──────┬────────────┤
    │             │            │
    ▼             ▼            ▼
┌────────┐ ┌──────────┐  ┌──────────┐
│ Vector │ │ LLM      │  │ Embeddings
│ DB     │ │ Provider │  │ Provider
│ Qdrant │ │ Ollama   │  │ Ollama
│ Chroma │ │ OpenAI   │  │ OpenAI
└────────┘ └──────────┘  └──────────┘

┌──────────────────────┐
│ Python Ingestor      │
│ • Load documents     │
│ • Chunk text         │
│ • Generate embeddings
│ • Store vectors      │
└──────────┬───────────┘
           │
           ▼
   ┌─────────────┐
   │ Data Files  │
   │ • PDFs      │
   │ • Markdown  │
   │ • CSV       │
   └─────────────┘
```

## ✨ Features

- **Multi-Format Document Support**
  - PDF, Markdown, Text, CSV, JSON, HTML
  - Recursive directory loading
  - Automatic metadata extraction

- **Intelligent Chunking**
  - Recursive chunking with overlap
  - Markdown-aware processing
  - Configurable chunk size and overlap

- **Flexible Retrieval**
  - Similarity search
  - Maximal Marginal Relevance (MMR)
  - Metadata filtering
  - Configurable score thresholds

- **Multiple LLM Providers**
  - Ollama (local inference)
  - OpenAI
  - Azure OpenAI
  - Claude
  - Gemini

- **Multiple Embedding Providers**
  - Ollama (local embeddings)
  - OpenAI
  - SentenceTransformers

- **Vector Database Support**
  - Qdrant (production-grade)
  - ChromaDB

- **Rich UI Features**
  - Real-time chat with streaming
  - Search interface
  - Source citations with page numbers
  - Dark mode
  - Conversation history
  - Confidence scores

- **Production-Ready**
  - Structured logging
  - Health checks
  - Metrics collection
  - Docker deployment
  - Graceful shutdown
  - Error handling

## 🚀 Quick Start

### Docker Compose (Recommended)

```bash
# Clone and navigate
cd /home/ajayp/code/AIProject/RAG/SparkRAG

# Start all services
docker-compose up -d

# Wait for services to initialize
sleep 30

# Access the UI
open http://localhost:3000
```

Pull Ollama models (optional, but recommended):
```bash
docker exec spark-rag-ollama ollama pull nomic-embed-text
docker exec spark-rag-ollama ollama pull mistral
```

### Local Development

```bash
# Terminal 1: Start infrastructure
docker-compose up -d qdrant ollama

# Terminal 2: Start Go API
cd api
go run .

# Terminal 3: Start React UI
cd ui
npm start

# Access at http://localhost:3000
```

## 📦 Installation

### Prerequisites

- Docker & Docker Compose 2.0+
- Go 1.21+
- Python 3.11+
- Node.js 18+
- 8GB RAM minimum (16GB recommended)

### From Source

```bash
# Clone repository
git clone <repository-url>
cd spark-rag-platform

# Initialize project
make init

# Build all components
make build

# Start with Docker Compose
make compose-up

# View logs
make compose-logs
```

### Environment Setup

1. Copy `.env.example` to `.env`
2. Configure your providers:
   ```bash
   cp .env.example .env
   # Edit .env with your API keys
   ```

3. Copy and customize config:
   ```bash
   cp config.yaml config.local.yaml
   # Edit configuration as needed
   ```

## 🔧 Configuration

### Via config.yaml

```yaml
server:
  port: 8080
  host: "0.0.0.0"

vectordb:
  provider: "qdrant"
  url: "http://localhost:6333"
  collection: "spark-rag"

embedding:
  provider: "ollama"
  model: "nomic-embed-text"

llm:
  provider: "ollama"
  url: "http://localhost:11434"
  model: "mistral"
  temperature: 0.7

retriever:
  top_k: 5
  chunk_size: 512
  chunk_overlap: 50
  strategy: "similarity"
  score_threshold: 0.5

logging:
  level: "info"
  json: false
```

### Via Environment Variables

```bash
VECTORDB_PROVIDER=qdrant
VECTORDB_URL=http://localhost:6333
EMBEDDING_PROVIDER=ollama
EMBEDDING_MODEL=nomic-embed-text
LLM_PROVIDER=ollama
LLM_URL=http://localhost:11434
LLM_MODEL=mistral
SERVER_PORT=8080
LOG_LEVEL=info
```

### Switching Vector Databases

**Qdrant (Production):**
```yaml
vectordb:
  provider: "qdrant"
  url: "http://localhost:6333"
  collection: "spark-rag"
```

**ChromaDB:**
```yaml
vectordb:
  provider: "chroma"
  url: "http://localhost:8000"
  collection: "spark-rag"
```

### Switching LLM Providers

**Ollama (Local):**
```yaml
llm:
  provider: "ollama"
  url: "http://localhost:11434"
  model: "mistral"
```

**OpenAI:**
```yaml
llm:
  provider: "openai"
  model: "gpt-4-turbo-preview"
  api_key: "sk-..."
  max_tokens: 2048
```

## 🐳 Docker Deployment

### Build Images

```bash
make docker
```

Or individually:
```bash
docker build -f Dockerfile.api -t spark-rag-api:latest .
docker build -f Dockerfile.ingestor -t spark-rag-ingestor:latest .
docker build -f Dockerfile.ui -t spark-rag-ui:latest .
```

### Run Services

```bash
# Full stack
docker-compose up -d

# Specific services
docker-compose up -d spark-rag-api spark-rag-ui

# With logs
docker-compose up

# Stop services
docker-compose down

# Remove volumes
docker-compose down -v
```

### Service Ports

- **UI**: http://localhost:3000
- **API**: http://localhost:8080
- **Qdrant**: http://localhost:6333
- **Ollama**: http://localhost:11434
- **ChromaDB**: http://localhost:8000 (optional)

## 📚 CLI Usage

### API Server

```bash
cd api

# Run with default config
go run . 

# Run with custom config
go run . -config=config.prod.yaml

# Override port
go run . -port=9000
```

### Ingestor

```bash
cd ingestor

# Ingest documents from data directory
python ingestor.py

# With custom config
python ingestor.py --config=config.yaml
```

## 📖 API Reference

All endpoints require `Content-Type: application/json`

### Chat

**POST /api/chat**

Request:
```json
{
  "message": "Explain Adaptive Query Execution in Spark",
  "conversation_id": "optional-id",
  "history": [
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ],
  "top_k": 5,
  "strategy": "similarity"
}
```

Response:
```json
{
  "conversation_id": "uuid",
  "answer": "AQE is a dynamic optimization technique...",
  "citations": [
    {
      "source": "Spark.pdf",
      "document": "spark-aqe",
      "page": 42,
      "score": 0.92
    }
  ],
  "confidence_score": 0.88,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Search

**POST /api/search**

Request:
```json
{
  "query": "Dynamic Partition Pruning",
  "top_k": 5,
  "strategy": "similarity"
}
```

Response:
```json
{
  "results": [
    {
      "document": "spark-optimization",
      "content": "...",
      "score": 0.89,
      "citation": {
        "source": "Spark.pdf",
        "page": 35
      }
    }
  ],
  "total": 1,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Health Check

**GET /health**

Response:
```json
{
  "status": "healthy",
  "services": {
    "vectordb": {"status": "healthy"},
    "llm": "ollama",
    "embedding": "ollama"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Providers

**GET /api/providers**

Response:
```json
{
  "embedding_providers": ["ollama", "openai"],
  "llm_providers": ["ollama", "openai"],
  "vectordb_providers": ["qdrant", "chroma"]
}
```

### Collections

**GET /api/collections**

Response:
```json
{
  "collections": ["spark-rag", "hadoop-docs"],
  "active": "spark-rag"
}
```

## 📁 Project Structure

```
spark-rag-platform/
├── api/                          # Go REST API
│   ├── main.go                   # Entry point
│   ├── config.go                 # Configuration
│   ├── types.go                  # Data types
│   ├── handlers.go               # HTTP handlers
│   ├── vectordb.go               # Vector DB abstraction
│   ├── embeddings.go             # Embedding providers
│   ├── llm.go                    # LLM providers
│   ├── retriever.go              # Retrieval logic
│   ├── logger.go                 # Logging setup
│   ├── go.mod                    # Dependencies
│   └── go.sum
│
├── ingestor/                     # Python document ingestion
│   ├── ingestor.py              # Main ingestor
│   ├── config.py                # Configuration
│   ├── document_loader.py       # Document parsing
│   ├── chunking.py              # Text chunking
│   ├── embeddings.py            # Embedding generation
│   ├── vectordb.py              # Vector DB client
│   ├── requirements.txt         # Python dependencies
│   └── tests/                   # Unit tests
│
├── ui/                           # React Chat UI
│   ├── public/
│   │   └── index.html           # HTML template
│   ├── src/
│   │   ├── index.js             # React entry
│   │   ├── index.css            # Global styles
│   │   ├── App.js               # Main component
│   │   ├── App.css              # App styles
│   │   ├── api.js               # API client
│   │   └── components/          # UI components
│   │       ├── ChatView.js
│   │       ├── SearchView.js
│   │       ├── SettingsView.js
│   │       ├── Message.js
│   │       └── Citation.js
│   └── package.json
│
├── data/                         # Sample documents
│   └── Spark.pdf                # Your existing PDF
│
├── docs/                         # Documentation
│   ├── ARCHITECTURE.md
│   ├── DEPLOYMENT.md
│   ├── CONTRIBUTING.md
│   └── API.md
│
├── docker-compose.yml           # Multi-container setup
├── Dockerfile.api               # API image
├── Dockerfile.ingestor          # Ingestor image
├── Dockerfile.ui                # UI image
├── Makefile                     # Build tasks
├── config.yaml                  # Configuration
├── .env                         # Environment variables
├── .env.example                 # Template
├── .gitignore
└── README.md
```

## 🔄 Adding Documents

### Supported Formats

- **PDF**: `.pdf`
- **Markdown**: `.md`
- **Text**: `.txt`
- **CSV**: `.csv`
- **JSON**: `.json`
- **HTML**: `.html`

### Ingestion

1. **Place files in data directory:**
   ```bash
   cp document.pdf data/
   cp guide.md data/
   ```

2. **Run ingestor:**
   ```bash
   docker-compose up spark-rag-ingestor
   # or
   python ingestor/ingestor.py
   ```

3. **Query immediately:**
   - The chat will include newly ingested documents
   - Search indexes update in real-time

### Metadata Extraction

The ingestor automatically extracts:
- Filename
- Document type
- Last modified date
- Page numbers (PDF)
- CSV row count
- Language detection

## 🧪 Development

### Running Tests

```bash
# Go tests
cd api
go test -v -cover ./...

# Python tests
cd ingestor
python -m pytest tests/ -v --cov

# Format code
make fmt
```

### Making Changes

1. **API Changes:**
   - Edit `api/*.go`
   - Test: `cd api && go test ./...`
   - Rebuild: `docker build -f Dockerfile.api .`

2. **Ingestor Changes:**
   - Edit `ingestor/*.py`
   - Test: `cd ingestor && pytest tests/`

3. **UI Changes:**
   - Edit `ui/src/**/*.js`
   - Hot reload in dev: `cd ui && npm start`

### Adding New LLM Providers

1. Implement `LLMProvider` interface in `api/llm.go`
2. Add provider detection in `main.go`
3. Test with custom config

### Adding New Embedding Models

1. Implement `EmbeddingProvider` in `api/embeddings.go`
2. Update ingestor in `ingestor/embeddings.py`
3. Configure via `config.yaml`

## 🚨 Troubleshooting

### Services won't start

```bash
# Check logs
docker-compose logs -f

# Restart services
docker-compose restart

# Full reset
docker-compose down -v
docker-compose up -d
```

### Connection refused errors

- Ensure services are healthy: `docker-compose ps`
- Check firewall/port conflicts: `lsof -i :8080`
- Verify network: `docker network ls`

### Poor search results

1. **Adjust retriever settings:**
   ```yaml
   retriever:
     top_k: 10        # Increase from 5
     score_threshold: 0.3  # Lower threshold
     strategy: "mmr"  # Try MMR
   ```

2. **Reindex documents:**
   ```bash
   rm ingestor/.ingestor.json
   docker-compose restart spark-rag-ingestor
   ```

3. **Check embeddings:**
   - Verify embedding model: `EMBEDDING_MODEL=nomic-embed-text`
   - Ensure Ollama is running: `docker exec spark-rag-ollama ollama list`

### UI not loading

- Check browser console for errors
- Verify API is running: `curl http://localhost:8080/health`
- Check CORS settings in `api/main.go`
- Clear browser cache

### Out of memory

- Increase Docker memory: `docker update --memory 4g spark-rag-ollama`
- Use smaller LLM: `LLM_MODEL=neural-chat` (instead of mistral)
- Reduce chunk size: `CHUNK_SIZE=256`

### API slow responses

- Check Qdrant: `curl http://localhost:6333/health`
- Monitor queries: `docker logs spark-rag-api`
- Reduce `top_k` in config
- Enable caching (future enhancement)

## 📈 Performance Tuning

### For Production

```yaml
llm:
  temperature: 0.3  # More deterministic

retriever:
  top_k: 3          # Fewer, better results
  reranker_enabled: true
  reranker_model: "cross-encoder/ms-marco-minilm-l-12-v2"
  strategy: "mmr"   # Better diversity

server:
  # Use reverse proxy (nginx)
  # Enable caching layer
```

### Docker Resource Limits

```yaml
services:
  ollama:
    deploy:
      resources:
        limits:
          memory: 8G
          cpus: '4'
```

## 📝 License

Apache 2.0

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -am 'Add feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Submit Pull Request

## 📞 Support

- Issues: GitHub Issues
- Discussions: GitHub Discussions
- Documentation: See `docs/` directory

---

**Built with ❤️ for the Apache Spark community**
