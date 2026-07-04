package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// EmbeddingProvider is the interface for embedding generation services.
// Implementations must handle both single and batch embeddings.
type EmbeddingProvider interface {
	// Embed generates an embedding for a single text string.
	// Returns the embedding vector or an error.
	Embed(ctx context.Context, text string) ([]float32, error)
	
	// EmbedBatch generates embeddings for multiple texts efficiently.
	// Should batch requests when possible for better performance.
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

// OllamaEmbedding implements EmbeddingProvider using local Ollama service.
// Ollama must be running and the embedding model pulled.
// Supported models: nomic-embed-text, all-minilm, and others.
type OllamaEmbedding struct {
	baseURL string      // Ollama API base URL
	model   string      // Model name (e.g., nomic-embed-text)
	client  *http.Client
}

// NewOllamaEmbedding creates an Ollama embedding provider.
// baseURL should be the Ollama API endpoint (default: http://localhost:11434)
// model is the embedding model name in Ollama
func NewOllamaEmbedding(baseURL, model string) *OllamaEmbedding {
	return &OllamaEmbedding{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// Embed generates a single embedding using Ollama API
func (o *OllamaEmbedding) Embed(ctx context.Context, text string) ([]float32, error) {
	payload := map[string]interface{}{
		"model":  o.model,
		"prompt": text,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/embeddings", o.baseURL), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embedding failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode embedding: %w", err)
	}

	return result.Embedding, nil
}

// EmbedBatch generates embeddings for multiple texts sequentially
func (o *OllamaEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	var embeddings [][]float32

	for _, text := range texts {
		emb, err := o.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings = append(embeddings, emb)
	}

	return embeddings, nil
}

// OpenAIEmbedding implements EmbeddingProvider using OpenAI cloud API.
// Supports embedding models: text-embedding-3-small, text-embedding-3-large
type OpenAIEmbedding struct {
	apiKey string      // OpenAI API key
	model  string      // Model name
	client *http.Client
}

// NewOpenAIEmbedding creates an OpenAI embedding provider
func NewOpenAIEmbedding(apiKey, model string) *OpenAIEmbedding {
	return &OpenAIEmbedding{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Embed generates a single embedding using OpenAI API
func (o *OpenAIEmbedding) Embed(ctx context.Context, text string) ([]float32, error) {
	payload := map[string]interface{}{
		"input": text,
		"model": o.model,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embedding failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode embedding: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return result.Data[0].Embedding, nil
}

// EmbedBatch generates embeddings for multiple texts in a single API call
func (o *OpenAIEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	payload := map[string]interface{}{
		"input": texts,
		"model": o.model,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embedding failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode embedding: %w", err)
	}

	// Sort by index to maintain order
	embeddings := make([][]float32, len(result.Data))
	for _, item := range result.Data {
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// HuggingFaceEmbedding implements EmbeddingProvider using HuggingFace Inference API.
// Supports any embedding model from the HuggingFace hub.
type HuggingFaceEmbedding struct {
	apiKey string      // HuggingFace API token
	model  string      // Model name
	apiURL string      // Full API endpoint URL
	client *http.Client
}

// NewHuggingFaceEmbedding creates a HuggingFace Inference API embedding provider
func NewHuggingFaceEmbedding(apiKey, model string) *HuggingFaceEmbedding {
	return &HuggingFaceEmbedding{
		apiKey: apiKey,
		model:  model,
		apiURL: fmt.Sprintf("https://api-inference.huggingface.co/models/%s", model),
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Embed generates a single embedding using HuggingFace Inference API
func (h *HuggingFaceEmbedding) Embed(ctx context.Context, text string) ([]float32, error) {
	payload := map[string]interface{}{
		"inputs": text,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", h.apiURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.apiKey))

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("huggingface embedding failed: %w", err)
	}
	defer resp.Body.Close()

	var result [][]float32
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode embedding: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return result[0], nil
}

// EmbedBatch generates embeddings for multiple texts using HuggingFace API
func (h *HuggingFaceEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	payload := map[string]interface{}{
		"inputs": texts,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", h.apiURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.apiKey))

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("huggingface embedding failed: %w", err)
	}
	defer resp.Body.Close()

	var result [][]float32
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode embedding: %w", err)
	}

	return result, nil
}
