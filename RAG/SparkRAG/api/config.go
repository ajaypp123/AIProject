package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int    `yaml:"port" envconfig:"SERVER_PORT"`
	Host string `yaml:"host" envconfig:"SERVER_HOST"`
}

type VectorDBConfig struct {
	Provider   string `yaml:"provider" envconfig:"VECTORDB_PROVIDER"`
	URL        string `yaml:"url" envconfig:"VECTORDB_URL"`
	APIKey     string `yaml:"api_key" envconfig:"VECTORDB_API_KEY"`
	Collection string `yaml:"collection" envconfig:"VECTORDB_COLLECTION"`
	BatchSize  int    `yaml:"batch_size" envconfig:"VECTORDB_BATCH_SIZE"`
}

type EmbeddingConfig struct {
	Provider  string `yaml:"provider" envconfig:"EMBEDDING_PROVIDER"`
	Model     string `yaml:"model" envconfig:"EMBEDDING_MODEL"`
	APIKey    string `yaml:"api_key" envconfig:"EMBEDDING_API_KEY"`
	URL       string `yaml:"url" envconfig:"EMBEDDING_URL"`
	BatchSize int    `yaml:"batch_size" envconfig:"EMBEDDING_BATCH_SIZE"`
}

type LLMConfig struct {
	Provider    string  `yaml:"provider" envconfig:"LLM_PROVIDER"`
	URL         string  `yaml:"url" envconfig:"LLM_URL"`
	Model       string  `yaml:"model" envconfig:"LLM_MODEL"`
	APIKey      string  `yaml:"api_key" envconfig:"LLM_API_KEY"`
	MaxTokens   int     `yaml:"max_tokens" envconfig:"LLM_MAX_TOKENS"`
	Temperature float32 `yaml:"temperature" envconfig:"LLM_TEMPERATURE"`
}

type RetrieverConfig struct {
	TopK            int     `yaml:"top_k" envconfig:"RETRIEVER_TOP_K"`
	ChunkSize       int     `yaml:"chunk_size" envconfig:"RETRIEVER_CHUNK_SIZE"`
	ChunkOverlap    int     `yaml:"chunk_overlap" envconfig:"RETRIEVER_CHUNK_OVERLAP"`
	RerankerEnabled bool    `yaml:"reranker_enabled" envconfig:"RETRIEVER_RERANKER_ENABLED"`
	RerankerModel   string  `yaml:"reranker_model" envconfig:"RETRIEVER_RERANKER_MODEL"`
	Strategy        string  `yaml:"strategy" envconfig:"RETRIEVER_STRATEGY"`
	ScoreThreshold  float32 `yaml:"score_threshold" envconfig:"RETRIEVER_SCORE_THRESHOLD"`
}

type LoggingConfig struct {
	Level string `yaml:"level" envconfig:"LOG_LEVEL"`
	JSON  bool   `yaml:"json" envconfig:"LOG_JSON"`
}

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	VectorDB  VectorDBConfig  `yaml:"vectordb"`
	Embedding EmbeddingConfig `yaml:"embedding"`
	LLM       LLMConfig       `yaml:"llm"`
	Retriever RetrieverConfig `yaml:"chunking"`
	Logging   LoggingConfig   `yaml:"logging"`
}

func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		VectorDB: VectorDBConfig{
			Provider:   "qdrant",
			URL:        "http://localhost:6333",
			Collection: "spark-rag",
		},
		Embedding: EmbeddingConfig{
			Provider: "ollama",
			Model:    "nomic-embed-text",
			URL:      "",
			APIKey:   "",
		},
		LLM: LLMConfig{
			Provider:    "ollama",
			URL:         "http://localhost:11434",
			Model:       "mistral",
			MaxTokens:   2048,
			Temperature: 0.7,
		},
		Retriever: RetrieverConfig{
			TopK:            5,
			ChunkSize:       512,
			ChunkOverlap:    50,
			RerankerEnabled: false,
			Strategy:        "similarity",
			ScoreThreshold:  0.5,
		},
		Logging: LoggingConfig{
			Level: "info",
			JSON:  false,
		},
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	return cfg, nil
}
