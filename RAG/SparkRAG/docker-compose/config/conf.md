
# VectorDB Configuration

## CromaDB Local Client
```yaml
vectordb:
  provider: "chroma-local"
  url: "/tmp/chroma_db"
  api_key: ""
  collection: "spark-rag"
  batch_size: 100
```

## CromaDB Url Client

```sh
docker pull chromadb/chroma:latest
docker run --rm -p 6333:8000 chromadb/chroma:latest
curl http://localhost:6333/api/v2/heartbeat
```

- `http://localhost:6333`

```yaml
vectordb:
  provider: "chroma"
  url: "http://localhost:6333"
  api_key: ""
  collection: "spark-rag"
  batch_size: 100
```

## Qdrant Client

```sh
docker run --rm -p 6333:6333 -p 6334:6334 \
  -v "$PWD/qdrant_storage:/qdrant/storage" \
  qdrant/qdrant:latest
curl http://localhost:6333/
```

```yaml
vectordb:
  provider: "qdrant"
  url: "http://localhost:6333"
  api_key: ""
  collection: "spark-rag"
  batch_size: 100
```

# Embedding Configuration

## sentence-transformers
```yaml
embedding:
  provider: "sentence-transformers"
  model: "sentence-transformers/all-MiniLM-L6-v2"
  api_key: ""
  url: ""
  batch_size: 32
```

## ollama

```sh
ollama pull nomic-embed-text
ollama list
```

```yaml
embedding:
  provider: "ollama"
  model: "nomic-embed-text"
  api_key: ""
  url: "http://localhost:11434"
  batch_size: 32
```

## OpenAI emmbedding models

```yaml
embedding:
  provider: "openai"
  model: ""
  api_key: ""
  url: ""
  batch_size: 32
```

# LLM

| Model         | Size | Speed | Quality | Recommendation          |
| ------------- | ---- | ----- | ------- | ----------------------- |
| `llama3.2:3b` | 3B   | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐    | Very fast testing       |
| `gemma3:4b`   | 4B   | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐   | Excellent small model   |
| `qwen2.5:7b`  | 7B   | ⭐⭐⭐⭐  | ⭐⭐⭐⭐⭐   | **Best overall choice** |
| `mistral`     | 7B   | ⭐⭐⭐⭐  | ⭐⭐⭐⭐    | Good alternative        |
| `qwen2.5:14b` | 14B  | ⭐⭐    | ⭐⭐⭐⭐⭐   | Too large for 8 GB VRAM |

```yaml
llm:
  provider: "ollama"
  url: "http://localhost:11434"
  model: "mistral"
  api_key: ""
  max_tokens: 2048
  temperature: 0.7
```
