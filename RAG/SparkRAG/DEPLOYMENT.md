# Spark RAG - Deployment Guide

## Production Deployment Checklist

### Pre-Deployment

- [ ] Review `.env` file - no hardcoded secrets
- [ ] Configure external LLM and embedding providers
- [ ] Set up persistent volumes for Qdrant
- [ ] Plan resource allocation
- [ ] Set up monitoring and logging
- [ ] Prepare SSL certificates
- [ ] Configure backups

### Configuration

#### 1. Environment Variables

```bash
# Database
VECTORDB_PROVIDER=qdrant
VECTORDB_URL=https://qdrant.example.com
VECTORDB_API_KEY=<secure-api-key>
VECTORDB_COLLECTION=spark-rag

# LLM Provider
LLM_PROVIDER=openai
LLM_API_KEY=sk-<secret>
LLM_MODEL=gpt-4-turbo-preview

# Embeddings
EMBEDDING_PROVIDER=openai
EMBEDDING_API_KEY=sk-<secret>
EMBEDDING_MODEL=text-embedding-3-large

# Server
SERVER_PORT=8080
LOG_LEVEL=info
LOG_JSON=true
```

#### 2. Docker Compose for Production

```yaml
version: '3.9'

services:
  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
    volumes:
      - qdrant-storage:/qdrant/storage
    environment:
      QDRANT_API_KEY: ${QDRANT_API_KEY}
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:6333/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  spark-rag-api:
    image: spark-rag-api:latest
    ports:
      - "8080:8080"
    environment:
      SERVER_HOST: 0.0.0.0
      SERVER_PORT: 8080
      LOG_LEVEL: info
      VECTORDB_URL: http://qdrant:6333
    depends_on:
      qdrant:
        condition: service_healthy
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 1G

  spark-rag-ingestor:
    image: spark-rag-ingestor:latest
    environment:
      VECTORDB_URL: http://qdrant:6333
      INGESTOR_WATCH_DIR: /data
    volumes:
      - ./data:/data:ro
      - ingestor-checkpoint:/app/checkpoint
    depends_on:
      qdrant:
        condition: service_healthy
    restart: on-failure
    deploy:
      resources:
        limits:
          memory: 1G

  spark-rag-ui:
    image: spark-rag-ui:latest
    ports:
      - "3000:3000"
    environment:
      REACT_APP_API_URL: http://localhost:8080/api
    depends_on:
      - spark-rag-api
    restart: always

volumes:
  qdrant-storage:
    driver: local
  ingestor-checkpoint:
    driver: local
```

### 3. Nginx Reverse Proxy

```nginx
upstream spark_api {
    server spark-rag-api:8080;
}

upstream spark_ui {
    server spark-rag-ui:3000;
}

server {
    listen 80;
    server_name spark-rag.example.com;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name spark-rag.example.com;

    ssl_certificate /etc/ssl/certs/spark-rag.crt;
    ssl_certificate_key /etc/ssl/private/spark-rag.key;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;

    # UI
    location / {
        proxy_pass http://spark_ui;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # API
    location /api/ {
        proxy_pass http://spark_api/api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Increase timeouts for long-running requests
        proxy_connect_timeout 600s;
        proxy_send_timeout 600s;
        proxy_read_timeout 600s;
    }

    # Health check
    location /health {
        proxy_pass http://spark_api/health;
    }
}
```

### 4. Kubernetes Deployment

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: spark-rag

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: spark-rag-config
  namespace: spark-rag
data:
  LOG_LEVEL: "info"
  LOG_JSON: "true"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: spark-rag-api
  namespace: spark-rag
spec:
  replicas: 3
  selector:
    matchLabels:
      app: spark-rag-api
  template:
    metadata:
      labels:
        app: spark-rag-api
    spec:
      containers:
      - name: api
        image: spark-rag-api:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: spark-rag-config
        env:
        - name: VECTORDB_URL
          value: "http://qdrant:6333"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: spark-rag-api
  namespace: spark-rag
spec:
  selector:
    app: spark-rag-api
  type: LoadBalancer
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
```

## Monitoring

### Health Checks

```bash
# API Health
curl http://localhost:8080/health

# Qdrant Health
curl http://localhost:6333/health

# Metrics
curl http://localhost:8080/metrics
```

### Logging

Set `LOG_JSON: "true"` for structured logging compatible with centralized log collection:

```bash
docker-compose logs --follow spark-rag-api | jq .
```

### Docker Stats

```bash
docker stats spark-rag-api spark-rag-ingestor spark-rag-ui qdrant
```

## Scaling

### Horizontal Scaling (API)

- Run multiple instances of `spark-rag-api`
- Use load balancer (nginx, HAProxy, cloud provider)
- All instances share same Qdrant instance
- Stateless design allows easy scaling

### Vertical Scaling

Increase Docker resource limits:

```yaml
deploy:
  resources:
    limits:
      memory: 4G
      cpus: '4'
```

### Database Optimization

```bash
# Create Qdrant cluster for HA
docker run -d -p 6333:6333 qdrant/qdrant:latest \
  -c /qdrant/qdrant_storage/qdrant.yaml
```

## Backup & Recovery

### Qdrant Backup

```bash
# Backup
docker exec spark-rag-qdrant \
  curl -X POST http://localhost:6333/snapshots

# Restore
docker cp spark-rag-qdrant:/qdrant/storage ./backup/
```

### Document Recovery

Keep documents in version control or S3:

```bash
# Sync to S3
aws s3 sync ./data s3://bucket-name/spark-rag/data/

# Re-ingest after recovery
docker-compose up spark-rag-ingestor
```

## Troubleshooting

### API slow responses

```bash
# Check Qdrant stats
curl http://localhost:6333/collections/spark-rag

# Check resource usage
docker stats spark-rag-api

# Increase timeout
proxy_read_timeout 120s;
```

### Memory issues

```bash
# Reduce chunk size
CHUNK_SIZE=256

# Reduce batch embeddings
EMBEDDING_BATCH_SIZE=16

# Use smaller model
LLM_MODEL=neural-chat
```

### Ingestion hangs

```bash
# Check logs
docker logs spark-rag-ingestor

# Restart ingestor
docker-compose restart spark-rag-ingestor

# Clear checkpoint
docker exec spark-rag-ingestor rm -f .ingestor.json
```

## Security Best Practices

1. **Secrets Management**
   - Use secret managers (Vault, AWS Secrets)
   - Don't commit `.env` to version control
   - Rotate API keys regularly

2. **Network Security**
   - Run behind VPN/firewall
   - Use HTTPS with valid certificates
   - Enable authentication/authorization

3. **Data Security**
   - Enable Qdrant authentication
   - Encrypt data at rest
   - Regular security audits

4. **API Rate Limiting**
   ```nginx
   limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
   limit_req zone=api_limit burst=20 nodelay;
   ```

## Performance Benchmarks

- Single API instance: ~100 req/s
- With Qdrant clustering: ~1000+ req/s
- Ingestion: ~10 documents/minute (with embeddings)
- Response time: 0.5-3s (depends on LLM)

## Support & Updates

- Check releases: https://github.com/spark-rag/platform
- Security updates: Enable auto-updates or regular checks
- Community support: GitHub Discussions
