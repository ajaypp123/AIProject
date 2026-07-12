package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var (
	configFile = flag.String("config", "", "Path to configuration file")
	port       = flag.Int("port", 0, "Port to listen on (overrides config)")
	version    = "1.0.0"
)

func main() {
	flag.Parse()

	if *configFile == "" {
		// Check env RAG_API_CFG_FILE
		if envConfigFile := os.Getenv("RAG_API_CFG_FILE"); envConfigFile != "" {
			*configFile = envConfigFile
		} else {
			*configFile = "config.yaml"
		}
	}

	cfg, err := LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if *port > 0 {
		cfg.Server.Port = *port
	}

	logger, err := NewLogger(cfg.Logging.Level, cfg.Logging.JSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Spark RAG API starting",
		zap.String("version", version),
		zap.Int("port", cfg.Server.Port),
		zap.String("vectordb", cfg.VectorDB.Provider),
		zap.String("llm", cfg.LLM.Provider),
		zap.String("config", *configFile),
	)

	var vectorDB VectorDB
	switch cfg.VectorDB.Provider {
	case "qdrant":
		vectorDB = NewQdrantVectorDB(cfg.VectorDB.URL, cfg.VectorDB.Collection,
			cfg.VectorDB.APIKey, cfg.VectorDB.BatchSize)
		logger.Info("Using Qdrant vector database", zap.String("url", cfg.VectorDB.URL))
	case "chroma":
		vectorDB = NewChromaVectorDB(cfg.VectorDB.URL, cfg.VectorDB.Collection,
			cfg.VectorDB.APIKey, cfg.VectorDB.BatchSize)
		logger.Info("Using Chroma vector database", zap.String("url", cfg.VectorDB.URL))
	default:
		logger.Fatal("unsupported vectordb provider", zap.String("provider", cfg.VectorDB.Provider))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := vectorDB.Health(ctx)
	if err != nil {
		logger.Warn("vector database health check failed", zap.Error(err))
	} else {
		logger.Info("Vector database health check passed", zap.Any("health", health))
	}

	var embedding EmbeddingProvider
	switch cfg.Embedding.Provider {
	case "ollama":
		embedding = NewOllamaEmbedding(cfg.LLM.URL, cfg.Embedding.Model)
		logger.Info("Using Ollama embeddings", zap.String("model", cfg.Embedding.Model))
	case "openai":
		embedding = NewOpenAIEmbedding(cfg.Embedding.APIKey, cfg.Embedding.Model)
		logger.Info("Using OpenAI embeddings", zap.String("model", cfg.Embedding.Model))
	case "sentence-transformers":
		baseURL := cfg.Embedding.URL
		if baseURL == "" {
			baseURL = "http://localhost:8000"
		}
		embedding = NewSentenceTransformerEmbedding(baseURL, cfg.Embedding.Model)
		logger.Info("Using SentenceTransformer embeddings",
			zap.String("model", cfg.Embedding.Model),
			zap.String("url", baseURL),
			zap.String("note", "Ensure embedding service is running"))
	default:
		logger.Fatal("unsupported embedding provider", zap.String("provider", cfg.Embedding.Provider))
	}

	var llm LLMProvider
	switch cfg.LLM.Provider {
	case "ollama":
		llm = NewOllamaLLM(cfg.LLM.URL, cfg.LLM.Model, cfg.LLM.MaxTokens, cfg.LLM.Temperature)
		logger.Info("Using Ollama LLM", zap.String("model", cfg.LLM.Model))
	case "openai":
		llm = NewOpenAILLM(cfg.LLM.APIKey, cfg.LLM.Model, cfg.LLM.MaxTokens, cfg.LLM.Temperature)
		logger.Info("Using OpenAI LLM", zap.String("model", cfg.LLM.Model))
	default:
		logger.Fatal("unsupported llm provider", zap.String("provider", cfg.LLM.Provider))
	}

	retriever := NewRetriever(vectorDB, embedding, cfg.Retriever)
	handler := NewHandler(retriever, vectorDB, llm, embedding, cfg, logger)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(corsMiddleware)

	router.Post("/api/chat", handler.Chat)
	router.Get("/api/chat/stream", handler.ChatStream)
	router.Post("/api/search", handler.Search)
	router.Post("/api/delete", handler.Delete)
	router.Get("/api/documents", handler.Documents)
	router.Get("/api/providers", handler.Providers)
	router.Get("/api/collections", handler.Collections)
	router.Get("/api/health", handler.Health)
	router.Get("/api/metrics", handler.Metrics)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"name":"Spark RAG API","version":"%s","status":"running"}`, version)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("API server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	sig := <-sigChan
	logger.Info("received signal", zap.String("signal", sig.String()))

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}

	logger.Info("API server stopped")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
