
# VectorDB Configuration

## CromaDB
```yaml
vectordb:
  provider: "chroma-local"
  url: "/tmp/chroma_db"
  api_key: ""
  collection: "spark-rag"
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
```

## Qdrant Client

```sh
docker run --rm -p 6333:6333 -p 6334:6334 qdrant/qdrant:latest
curl http://localhost:6333/
```

```yaml
vectordb:
  provider: "qdrant"
  url: "http://localhost:6333"
  api_key: ""
  collection: "spark-rag"
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
```

```yaml
embedding:
  provider: "ollama"
  model: "nomic-embed-text"
  api_key: ""
  url: "http://localhost:11434"
  batch_size: 32
```
