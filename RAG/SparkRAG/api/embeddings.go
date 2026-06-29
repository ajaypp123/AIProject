package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

type OllamaEmbedding struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaEmbedding(baseURL, model string) *OllamaEmbedding {
	return &OllamaEmbedding{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (o *OllamaEmbedding) Embed(ctx context.Context, text string) ([]float32, error) {
	payload := map[string]interface{}{
		"model": o.model,
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

type OpenAIEmbedding struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIEmbedding(apiKey, model string) *OpenAIEmbedding {
	return &OpenAIEmbedding{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

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
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode embedding: %w", err)
	}

	var embeddings [][]float32
	for _, item := range result.Data {
		embeddings = append(embeddings, item.Embedding)
	}

	return embeddings, nil
}
