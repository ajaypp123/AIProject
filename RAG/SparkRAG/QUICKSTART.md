# Spark RAG - Quick Start Guide

Get the Spark RAG Platform running in 5 minutes!

## Prerequisites

- Docker & Docker Compose installed
- 8GB RAM minimum
- Internet connection (first run downloads models)

## Step 1: Clone and Navigate

```bash
cd /home/ajayp/code/AIProject/RAG/SparkRAG
```

## Step 2: Start Services

```bash
docker-compose up -d
```

This starts:
- **Qdrant** (Vector Database) - Port 6333
- **Ollama** (LLM & Embeddings) - Port 11434
- **API** (Go REST API) - Port 8080
- **Ingestor** (Python) - Processes documents
- **UI** (React) - Port 3000

Wait 30 seconds for services to initialize.

## Step 3: Access the UI

Open your browser: **http://localhost:3000**

## Step 4: Try It Out

Type a question:
- "What is Adaptive Query Execution in Spark?"
- "Explain Delta Lake"
- "How does Qdrant work?"

The system will search your documents and generate answers!

## Common Issues

### Services won't start

```bash
# Check logs
docker-compose logs -f

# Restart
docker-compose restart
```

### Slow first response

Ollama is downloading AI models (~2-3GB). First request takes time.

```bash
# Check status
docker exec spark-rag-ollama ollama list
```

### Port already in use

Change ports in `docker-compose.yml`:

```yaml
ports:
  - "3001:3000"  # Change UI port to 3001
```

## Add More Documents

1. Place documents in `./data/` folder
2. Run ingestor:
   ```bash
   docker-compose restart spark-rag-ingestor
   ```
3. Chat immediately - documents are indexed!

## Stop Services

```bash
docker-compose down
```

## Use Different LLM

Edit `.env`:

```bash
# Use OpenAI instead of Ollama
LLM_PROVIDER=openai
LLM_API_KEY=sk-...
LLM_MODEL=gpt-4-turbo-preview

# Then restart
docker-compose restart spark-rag-api
```

## View API Documentation

```bash
curl http://localhost:8080/api/providers
curl http://localhost:8080/health
```

## Next Steps

- Read [README.md](./README.md) for full documentation
- Check [DEPLOYMENT.md](./DEPLOYMENT.md) for production setup
- Explore API in [api/handlers.go](./api/handlers.go)

---

**Questions?** Check the troubleshooting section in README.md
