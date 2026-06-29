.PHONY: help build run test lint fmt docker clean compose-up compose-down compose-logs init

help:
	@echo "Spark RAG Platform - Available targets:"
	@echo "  make init           - Initialize project"
	@echo "  make build          - Build all components"
	@echo "  make run            - Run locally (development)"
	@echo "  make test           - Run tests"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo "  make docker         - Build Docker images"
	@echo "  make compose-up     - Start docker-compose"
	@echo "  make compose-down   - Stop docker-compose"
	@echo "  make compose-logs   - View docker-compose logs"
	@echo "  make clean          - Clean build artifacts"

init:
	@echo "Initializing Spark RAG Platform..."
	mkdir -p data
	cd api && go mod tidy
	cd ingestor && pip install -r requirements.txt
	cd ui && npm install
	@echo "Initialization complete!"

build: build-api build-ingestor build-ui
	@echo "All components built successfully!"

build-api:
	@echo "Building Go API..."
	cd api && go build -o spark-rag-api .
	@echo "API built successfully!"

build-ingestor:
	@echo "Installing Python dependencies..."
	cd ingestor && pip install -r requirements.txt
	@echo "Ingestor ready!"

build-ui:
	@echo "Building React UI..."
	cd ui && npm run build
	@echo "UI built successfully!"

run:
	@echo "Starting Spark RAG Platform (development mode)..."
	docker-compose up -d qdrant ollama
	@echo "Infrastructure running. Start API in another terminal: cd api && go run ."
	@echo "And start UI: cd ui && npm start"

test:
	@echo "Running Go tests..."
	cd api && go test -v -cover ./...
	@echo "Running Python tests..."
	cd ingestor && python -m pytest tests/ -v --cov
	@echo "All tests passed!"

lint:
	@echo "Running linters..."
	cd api && go vet ./...
	cd ingestor && pylint ingestor/ || true
	@echo "Linting complete!"

fmt:
	@echo "Formatting code..."
	cd api && gofmt -s -w .
	cd ingestor && black . --quiet
	@echo "Code formatted!"

docker:
	@echo "Building Docker images..."
	docker build -f Dockerfile.api -t spark-rag-api:latest .
	docker build -f Dockerfile.ingestor -t spark-rag-ingestor:latest .
	docker build -f Dockerfile.ui -t spark-rag-ui:latest .
	@echo "Docker images built successfully!"

compose-up:
	@echo "Starting Docker Compose..."
	docker-compose up -d
	@echo "Services starting. Use 'make compose-logs' to view logs."
	@echo "UI: http://localhost:3000"
	@echo "API: http://localhost:8080"

compose-down:
	@echo "Stopping Docker Compose..."
	docker-compose down
	@echo "Services stopped!"

compose-logs:
	docker-compose logs -f

clean:
	@echo "Cleaning build artifacts..."
	rm -f api/spark-rag-api
	rm -rf ui/build
	rm -rf api/.env.local
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	@echo "Clean complete!"
