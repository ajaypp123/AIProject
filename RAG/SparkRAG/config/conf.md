
### VectorDBLocally: CromaDB
```yaml
vectordb:
  provider: "chroma-local"
  url: "/tmp/chroma_db"
  api_key: ""
  collection: "spark-rag"
```

### VectorDBDocker: CromaDB Docker

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
