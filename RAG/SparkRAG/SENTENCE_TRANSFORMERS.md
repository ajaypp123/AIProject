# Using SentenceTransformers Embeddings

This guide explains how to use local, CPU-efficient embeddings with the Spark RAG API using SentenceTransformers.

## Why SentenceTransformers?

- **Local processing**: No API keys or internet required
- **Fast**: CPU-efficient inference
- **Cost-free**: No per-request charges
- **Private**: Your data never leaves your machine
- **Flexible**: Supports many HuggingFace models

## Quick Start

### Option 1: Standalone (Development)

1. **Install dependencies:**
   ```bash
   cd ingestor
   pip install fastapi uvicorn sentence-transformers
   ```

2. **Start the embedding service:**
   ```bash
   python embedding_service.py --model sentence-transformers/all-MiniLM-L6-v2 --port 8000
   ```

   Service will be available at `http://localhost:8000`

3. **In another terminal, start the API:**
   ```bash
   cd api
   go run . -config ../config/config.development.yaml
   ```

4. **Configuration (already set in config.development.yaml):**
   ```yaml
   embedding:
     provider: "sentence-transformers"
     model: "sentence-transformers/all-MiniLM-L6-v2"
     url: "http://localhost:8000"  # or leave empty for default
     batch_size: 32
   ```

### Option 2: Docker Compose (Production)

1. **Start all services:**
   ```bash
   docker-compose up -d
   ```

2. **Check services:**
   ```bash
   # Embedding service health
   curl http://localhost:8001/health

   # API health
   curl http://localhost:8080/api/health
   ```

3. **Update config for Docker:**
   In your config, set embedding URL to service name:
   ```yaml
   embedding:
     url: "http://spark-rag-embedding-service:8000"
   ```

## Available Models

### Recommended (Fast & Good Quality)
- `sentence-transformers/all-MiniLM-L6-v2` (23MB, fastest)
- `sentence-transformers/all-distilroberta-v1` (268MB, very fast)

### High Quality
- `sentence-transformers/all-mpnet-base-v2` (430MB)
- `BAAI/bge-base-en-v1.5` (430MB, excellent)

### Best Quality (Slower)
- `BAAI/bge-large-en-v1.5` (1.2GB, best)
- `sentence-transformers/all-roberta-large-v1` (1.1GB)

### Change Model

```bash
# Stop current service (if running in Docker, see below)
python embedding_service.py --model sentence-transformers/all-mpnet-base-v2

# Or in docker-compose.yml:
services:
  spark-rag-embedding-service:
    environment:
      MODEL_NAME: sentence-transformers/all-mpnet-base-v2
```

## Monitoring

### Check service health
```bash
curl http://localhost:8000/health
```

Response:
```json
{
  "status": "healthy",
  "model": "sentence-transformers/all-MiniLM-L6-v2"
}
```

### View service logs
```bash
# Standalone
# Check terminal where you started embedding_service.py

# Docker
docker logs spark-rag-embedding-service -f
```

## Performance Tuning

### GPU Acceleration
Install CUDA-enabled PyTorch:
```bash
pip install torch[cuda] -f https://download.pytorch.org/whl/torch_stable.html
```

### Memory Usage
- Smaller models use less memory
- Batch size can be adjusted in config
- Default batch_size: 32

### CPU Optimization
For large batches:
```bash
python embedding_service.py --model sentence-transformers/all-MiniLM-L6-v2
# CPU will auto-parallelize across cores
```

## Troubleshooting

### Service won't start
```bash
# Check Python version (need 3.8+)
python --version

# Install/upgrade dependencies
pip install -U fastapi uvicorn sentence-transformers
```

### Connection refused
```bash
# Verify service is running
curl http://localhost:8000/health

# Check URL in Go config matches service address
# Default: http://localhost:8000 (standalone)
# Docker: http://spark-rag-embedding-service:8000
```

### Out of memory
```bash
# Use smaller model
python embedding_service.py --model sentence-transformers/all-MiniLM-L6-v2

# Or reduce batch size in config
# embedding:
#   batch_size: 16
```

### Model download fails
```bash
# Set cache directory
export HF_HOME=/path/to/cache
python embedding_service.py

# Or pre-download model
python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')"
```

## Migration from Ollama Embeddings

If you're currently using Ollama for embeddings, switch to SentenceTransformers:

**Before (config.yaml):**
```yaml
embedding:
  provider: "ollama"
  model: "nomic-embed-text"
  url: "http://localhost:11434"
```

**After (config.yaml):**
```yaml
embedding:
  provider: "sentence-transformers"
  model: "sentence-transformers/all-MiniLM-L6-v2"
  url: "http://localhost:8000"
```

Then:
1. Start embedding service: `python embedding_service.py`
2. Restart Go API: `go run . -config config.yaml`

No other changes needed!
