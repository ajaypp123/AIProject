# SentenceTransformer Embedding Service

Local HTTP service that wraps HuggingFace SentenceTransformers for use by the Go API.

This allows the Go API to use local, CPU-efficient embeddings without requiring external API keys or cloud services.

## Installation

```bash
pip install fastapi uvicorn sentence-transformers
```

## Usage

### Start the embedding service:

```bash
# With default model (all-MiniLM-L6-v2)
python embedding_service.py

# With custom model
python embedding_service.py --model sentence-transformers/all-mpnet-base-v2 --port 8000
```

### Configuration in API

In `config.development.yaml` or `config.production.yaml`:

```yaml
embedding:
  provider: "sentence-transformers"
  model: "sentence-transformers/all-MiniLM-L6-v2"
  url: "http://localhost:8000"
  api_key: ""
  batch_size: 32
```

The Go API will connect to this local service automatically.

## Supported Models

- `sentence-transformers/all-MiniLM-L6-v2` (fast, ~23MB)
- `sentence-transformers/all-mpnet-base-v2` (high quality, ~430MB)
- `BAAI/bge-base-en-v1.5` (excellent, ~430MB)
- `BAAI/bge-large-en-v1.5` (best quality, ~1.2GB)
- Any other SentenceTransformer model from HuggingFace Hub

## API Endpoints

### POST /embed
Generate embedding for a single text.

**Request:**
```json
{
  "text": "Your text here"
}
```

**Response:**
```json
{
  "embedding": [0.123, -0.456, ...]
}
```

### POST /embed-batch
Generate embeddings for multiple texts.

**Request:**
```json
{
  "texts": ["Text 1", "Text 2", "Text 3"]
}
```

**Response:**
```json
{
  "embeddings": [
    [0.123, -0.456, ...],
    [0.789, 0.321, ...],
    ...
  ]
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "model": "sentence-transformers/all-MiniLM-L6-v2"
}
```

## Docker Usage

### Build image:
```bash
docker build -f Dockerfile.embedding-service -t spark-rag-embedding-service .
```

### Run container:
```bash
docker run -p 8000:8000 \
  -e MODEL_NAME="sentence-transformers/all-MiniLM-L6-v2" \
  spark-rag-embedding-service
```

## Performance Notes

- First run will download and cache the model (~100MB-1.2GB depending on model)
- Models are cached in `~/.cache/huggingface/hub/`
- Batch processing is more efficient than individual requests
- CPU-based inference; GPU support via `pip install torch[cuda]`

## Troubleshooting

### Model download fails
```bash
# Set HuggingFace cache directory
export HF_HOME=/path/to/cache
python embedding_service.py
```

### Out of memory
Use a smaller model:
```bash
python embedding_service.py --model sentence-transformers/all-MiniLM-L6-v2
```

### Connection refused from Go API
- Verify service is running: `curl http://localhost:8000/health`
- Check `url` in Go config matches service address
- Ensure no firewall blocking port 8000
