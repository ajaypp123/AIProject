# SentenceTransformers Integration Summary

This implementation provides local, CPU-efficient embeddings for the Spark RAG system, matching the Python ingestor's capabilities.

## What Was Changed

### 1. **Go API (`api/embeddings.go`)**
- Added `SentenceTransformerEmbedding` struct
- Implements `EmbeddingProvider` interface
- Makes HTTP calls to local embedding service
- Supports both single and batch embeddings

### 2. **Go API (`api/main.go`)**
- Added case for `"sentence-transformers"` provider
- Connects to local service via HTTP
- Default URL: `http://localhost:8000`
- Logs service connection info

### 3. **Python Embedding Service** (NEW)
- **File**: `ingestor/embedding_service.py`
- FastAPI HTTP wrapper around SentenceTransformers
- Exposes `/embed` and `/embed-batch` endpoints
- Direct mirror of Python ingestor's `SentenceTransformerEmbedding` class
- Runs locally, no API keys needed

### 4. **Docker Support** (NEW)
- **File**: `Dockerfile.embedding-service`
- Builds standalone embedding service container
- **Updated**: `docker-compose.yml` with embedding service
- Service runs on port 8000 (8001 externally)

### 5. **Documentation** (NEW)
- **`SENTENCE_TRANSFORMERS.md`**: Complete usage guide
- **`ingestor/EMBEDDING_SERVICE.md`**: Service API documentation
- **This file**: Implementation details

## Architecture

```
Go API Server (8080)
    ↓
    └→ HTTP calls to
        ↓
Embedding Service (8000)
    ↓
    └→ SentenceTransformer Model (local)
```

## Configuration

### `config/config.development.yaml`
```yaml
embedding:
  provider: "sentence-transformers"
  model: "sentence-transformers/all-MiniLM-L6-v2"
  url: ""  # Defaults to http://localhost:8000
  batch_size: 32
```

## How It Works

### Python Side (Ingestor)
```python
class SentenceTransformerEmbedding(EmbeddingProvider):
    # Loads model directly
    from sentence_transformers import SentenceTransformer
    self.model = SentenceTransformer(model_name)

    # Generates embeddings locally
    embedding = self.model.encode(text)
```

### Go Side (API)
```go
type SentenceTransformerEmbedding struct {
    baseURL string  // "http://localhost:8000"
    model   string
    client  *http.Client
}

func (s *SentenceTransformerEmbedding) Embed() {
    // Makes HTTP POST to embedding service
    // Gets back embedding vector
}
```

### Embedding Service (HTTP Bridge)
```python
@app.post("/embed")
async def embed(request: EmbedRequest):
    # Receives text, generates embedding locally
    embedding = model.encode(request.text)
    return {"embedding": embedding.tolist()}
```

## Running the System

### Development (Standalone)

Terminal 1 - Start embedding service:
```bash
cd ingestor
pip install fastapi uvicorn sentence-transformers
python embedding_service.py
```

Terminal 2 - Start API:
```bash
cd api
go run . -config ../config/config.development.yaml
```

Terminal 3 - Start UI:
```bash
cd ui
npm start
```

### Production (Docker Compose)

```bash
docker-compose up -d
```

All services start automatically:
- Qdrant vector database
- Ollama (for LLM)
- Embedding service (local)
- Go API
- React UI

## Comparison: Python vs Go Implementation

| Aspect | Python Ingestor | Go API |
|--------|-----------------|--------|
| Embedding Model | Loaded directly in Python | Remote HTTP service |
| Initialization | `SentenceTransformer(model)` | `NewSentenceTransformerEmbedding(url, model)` |
| Single Embedding | `model.encode(text)` | HTTP POST to `/embed` |
| Batch Embeddings | `model.encode(texts)` | HTTP POST to `/embed-batch` |
| Dependencies | sentence-transformers | HTTP client only |
| Performance | Direct Python execution | Network I/O overhead |

## Advantages of This Design

1. **Separation of Concerns**: Embedding service is independent
2. **Language Agnostic**: Go doesn't need Python libraries
3. **Scalability**: Service can run on different machine
4. **Flexibility**: Can swap embedding implementations
5. **Debugging**: HTTP requests are easy to inspect
6. **Docker Ready**: Service containerizes easily

## Equivalent Functionality

Both Python and Go now support:
- ✅ Local CPU embeddings (no API keys)
- ✅ Multiple HuggingFace models
- ✅ Batch processing
- ✅ Same configuration format
- ✅ Docker deployment

## Next Steps

1. **Test locally**:
   ```bash
   python ingestor/embedding_service.py &
   cd api && go run . -config ../config/config.development.yaml
   ```

2. **Test batch embeddings**:
   ```bash
   curl -X POST http://localhost:8000/embed-batch \
     -H "Content-Type: application/json" \
     -d '{"texts": ["Hello", "World"]}'
   ```

3. **Test API**:
   ```bash
   curl http://localhost:8080/api/health
   ```

4. **Docker deployment**:
   ```bash
   docker-compose up -d
   ```

## Troubleshooting

See `SENTENCE_TRANSFORMERS.md` for detailed troubleshooting guide.

**Quick checks**:
```bash
# Service health
curl http://localhost:8000/health

# API health
curl http://localhost:8080/api/health

# View logs
docker logs spark-rag-embedding-service
```
