.PHONY: help build run test lint fmt docker clean compose-up compose-down compose-logs init

# help target to display available make targets
help:
	@echo "Spark RAG Platform - Available targets:"
	@echo "  make init           			- Initialize project"
	@echo "  make build          			- Build all components (ui, api, ingestor)"
	@echo "  make run-app-locally			- Run all component locally (development)"
	@echo "  make test           			- Run all component test cases"
	@echo "  make lint           			- Run linters"
	@echo "  make fmt            			- Format code"
	@echo "  make docker-build   			- Build Docker images for all components"
	@echo "  make run-docker-services     		- Start docker-compose services"
	@echo "  make stop-docker-services   		- Stop docker-compose services"
	@echo "  make show-component-logs   		- View docker-compose logs for all services"
	@echo "  make clean          			- Clean build artifacts and stop running services"

# init application, create volumes, install dependencies locally
init:
	@echo "Initializing Spark RAG Platform..."
	@echo "Make sure /data directory is writable by the current user."
	mkdir -p /data /data/ollama_storage /data/qdrant_storage /data/ingestor
	cp -rf ./data /data/ingestor/
	@echo "All documents for rag-ingestor are copied to /data/ingestor/data"
	ls /data/ingestor/data
	@echo "Initialization complete!"

# build all application components locally (Go API, Python Ingestor, React UI)
build: build-api build-ingestor build-ui
	@echo "All components built successfully!"

build-api:
	@echo "Building Go API..."
	cd api && go build -o ./bin/spark-rag-api .
	@echo "API built successfully!"

build-ingestor:
	@echo "Installing Python dependencies..."
	cd ingestor && pip install -r requirements.txt
	@echo "Ingestor ready!"

build-ui:
	@echo "Building React UI..."
	cd ui && npm install
	@echo "UI built successfully!"

# run-infra target to start Qdrant and Ollama services using Docker Compose
run-infra: init
	@echo "Starting Spark RAG Platform (development mode)..."
	cd ./docker-compose && docker compose -f ./docker-compose.yml up -d qdrant ollama
	@echo "Waiting for infrastructure to be ready..."
	# sleep 2m
	@echo "Infrastructure running."

# run the application locally (development mode)
run-app-locally: build run-ingestor run-rag-ui
	@echo "Spark RAG Platform is running locally (development mode)."

run-ingestor: run-infra
	@echo "Running Ingestor..."
	cd ingestor && python rag-ingestor_cli.py store --conf ../docker-compose/config/config.local.yaml
	@echo "Ingestor finished."

run-rag-ui: run-infra
	@echo ""
	@echo "Starting API..."
	@(cd api && ./bin/spark-rag-api --config ../docker-compose/config/config.local.yaml > api.log 2>&1 &)
	@echo "Starting UI..."
	@(cd ui && npm run dev > ui.log 2>&1 &)
	@echo "==============================="
	@echo "API : http://localhost:8080"
	@echo "UI  : http://localhost:3000"
	@echo "==============================="

# Perform tests for Go and Python code
test:
	@echo "Running Go tests..."
	cd api && go test -v -cover ./...
	@echo "Running Python tests..."
	pip install pytest pytest-cov
	cd ingestor && python -m pytest tests/ -v --cov
	@echo "All tests passed!"

# Run linters for Go and Python code
lint:
	@echo "Running linters..."
	cd api && go vet ./...
	cd ingestor && pylint ./ || true
	@echo "Linting complete!"

# Format code for Go and Python
fmt:
	@echo "Formatting code..."
	cd api && gofmt -s -w .
	cd ingestor && black . --quiet
	@echo "Code formatted!"

# Build Docker images for API, Ingestor, and UI
docker-build: docker-build-spark-rag-api docker-build-spark-rag-ingestor docker-build-spark-rag-ui
	@echo "All Docker images built successfully!"

docker-build-spark-rag-api:
	@echo "Building API Docker images..."
	docker build -f Dockerfile.api -t spark-rag-api:latest .
	@echo "All Docker images built successfully!"

docker-build-spark-rag-ingestor:
	@echo "Building Ingestor Docker images..."
	docker build -f Dockerfile.ingestor -t spark-rag-ingestor:latest .
	@echo "All Docker images built successfully!"

docker-build-spark-rag-ui:
	@echo "Building UI Docker images..."
	docker build -f Dockerfile.ui -t spark-rag-ui:latest .
	@echo "All Docker images built successfully!"

run-docker-services:
	@echo "Starting Docker Compose..."
	cd ./docker-compose && docker compose -f ./docker-compose.yml up -d
	@echo "Services starting. Use 'make compose-logs' to view logs."
	@echo "================================"
	@echo "UI: http://localhost:3000"
	@echo "API: http://localhost:8080"
	@echo "================================"

stop-docker-services:
	@echo "Stopping Docker Compose..."
	cd ./docker-compose && docker compose -f ./docker-compose.yml down -v
	@echo "Stopped all services."

show-component-logs:
	cd ./docker-compose && docker compose -f ./docker-compose.yml logs -f

clean:
	@echo "Cleaning build artifacts..."
	rm -rf ui/build ui/node_modules ui/package-lock.json
	rm -rf api/.env.local api/vendor/ api/bin/spark-rag-api
	rm -rf ingestor/.pytest_cache
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	pkill -f spark-rag-api
	fuser -k 3000/tcp
	@echo "Clean complete!"
