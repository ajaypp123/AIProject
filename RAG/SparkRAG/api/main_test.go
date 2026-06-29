package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}

	if cfg.VectorDB.Provider != "qdrant" {
		t.Errorf("Expected vectordb provider qdrant, got %s", cfg.VectorDB.Provider)
	}

	if cfg.LLM.Provider != "ollama" {
		t.Errorf("Expected llm provider ollama, got %s", cfg.LLM.Provider)
	}
}

func TestQdrantVectorDB(t *testing.T) {
	db := NewQdrantVectorDB("http://localhost:6333", "test-collection")

	if db == nil {
		t.Fatal("NewQdrantVectorDB() returned nil")
	}

	if db.baseURL != "http://localhost:6333" {
		t.Errorf("Expected baseURL http://localhost:6333, got %s", db.baseURL)
	}

	if db.collection != "test-collection" {
		t.Errorf("Expected collection test-collection, got %s", db.collection)
	}
}

func TestChromaVectorDB(t *testing.T) {
	db := NewChromaVectorDB("http://localhost:8000", "test-collection")

	if db == nil {
		t.Fatal("NewChromaVectorDB() returned nil")
	}

	if db.baseURL != "http://localhost:8000" {
		t.Errorf("Expected baseURL http://localhost:8000, got %s", db.baseURL)
	}
}

func TestOllamaEmbedding(t *testing.T) {
	embedding := NewOllamaEmbedding("http://localhost:11434", "nomic-embed-text")

	if embedding == nil {
		t.Fatal("NewOllamaEmbedding() returned nil")
	}

	if embedding.baseURL != "http://localhost:11434" {
		t.Errorf("Expected baseURL http://localhost:11434, got %s", embedding.baseURL)
	}

	if embedding.model != "nomic-embed-text" {
		t.Errorf("Expected model nomic-embed-text, got %s", embedding.model)
	}
}

func TestPromptBuilder(t *testing.T) {
	pb := NewPromptBuilder()

	if pb == nil {
		t.Fatal("NewPromptBuilder() returned nil")
	}

	if len(pb.systemPrompt) == 0 {
		t.Fatal("PromptBuilder has empty system prompt")
	}
}

func TestCitationBuilder(t *testing.T) {
	cb := NewCitationBuilder()

	if cb == nil {
		t.Fatal("NewCitationBuilder() returned nil")
	}

	results := []SearchResult{
		{
			Document: "doc1",
			Content:  "test content",
			Score:    0.95,
			Citation: Citation{
				Source: "test.pdf",
			},
		},
	}

	citations := cb.BuildCitations(results)
	if len(citations) != 1 {
		t.Errorf("Expected 1 citation, got %d", len(citations))
	}

	if citations[0].Source != "test.pdf" {
		t.Errorf("Expected source test.pdf, got %s", citations[0].Source)
	}
}
