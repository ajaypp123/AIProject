package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type VectorDB interface {
	Search(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]SearchResult, error)
	Store(ctx context.Context, doc Document) error
	Delete(ctx context.Context, docID string) error
	Health(ctx context.Context) (map[string]interface{}, error)
	Stats(ctx context.Context) (map[string]interface{}, error)
	Collections(ctx context.Context) ([]string, error)
}

type QdrantVectorDB struct {
	baseURL    string
	collection string
	httpClient *http.Client
	apiKey     string
	batchSize  int
}

func NewQdrantVectorDB(url, collection string, apiKey string, batchSize int) *QdrantVectorDB {
	return &QdrantVectorDB{
		baseURL:    url,
		collection: collection,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		batchSize:  batchSize,
	}
}

func (q *QdrantVectorDB) Search(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]SearchResult, error) {
	payload := map[string]interface{}{
		"vector":       embedding,
		"limit":        topK,
		"with_vectors": true,
		"with_payload": true,
	}

	if len(filters) > 0 {
		payload["filter"] = filters
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/collections/%s/points/search", q.baseURL, q.collection)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if q.apiKey != "" {
		req.Header.Set("api-key", fmt.Sprintf("%s", q.apiKey))
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qdrant search failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Result []struct {
			ID      interface{}            `json:"id"`
			Score   float32                `json:"score"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode qdrant response: %w", err)
	}

	var results []SearchResult
	for _, hit := range result.Result {
		sr := SearchResult{
			Score: hit.Score,
			Citation: Citation{
				Score: hit.Score,
			},
		}

		if doc, ok := hit.Payload["document"].(string); ok {
			sr.Document = doc
		}
		if content, ok := hit.Payload["content"].(string); ok {
			sr.Content = content
		}
		if source, ok := hit.Payload["source"].(string); ok {
			sr.Citation.Source = source
		}

		sr.Metadata = hit.Payload
		results = append(results, sr)
	}

	return results, nil
}

func (q *QdrantVectorDB) Store(ctx context.Context, doc Document) error {
	for _, chunk := range doc.Chunks {
		payload := map[string]interface{}{
			"document":      doc.ID,
			"title":         doc.Title,
			"content":       chunk.Content,
			"source":        doc.Source,
			"document_type": doc.DocumentType,
			"chunk_id":      chunk.ID,
		}

		point := map[string]interface{}{
			"id":      chunk.ID,
			"vector":  chunk.Embedding,
			"payload": payload,
		}

		body, _ := json.Marshal(map[string]interface{}{"points": []interface{}{point}})
		req, _ := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/collections/%s/points?wait=true", q.baseURL, q.collection), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		if q.apiKey != "" {
			req.Header.Set("api-key", fmt.Sprintf("%s", q.apiKey))
		}

		resp, err := q.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("qdrant store failed: %w", err)
		}
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("qdrant store returned status %d", resp.StatusCode)
		}
	}

	return nil
}

func (q *QdrantVectorDB) Delete(ctx context.Context, docID string) error {
	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"key": "document",
				"match": map[string]interface{}{
					"value": docID,
				},
			},
		},
	}

	body, _ := json.Marshal(filter)
	req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/collections/%s/points/delete", q.baseURL, q.collection), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if q.apiKey != "" {
		req.Header.Set("api-key", fmt.Sprintf("%s", q.apiKey))
	}

	resp, err := q.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant delete failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("qdrant delete returned status %d", resp.StatusCode)
	}

	return nil
}

func (q *QdrantVectorDB) Health(ctx context.Context) (map[string]interface{}, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s", q.baseURL), nil)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qdrant health check failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (q *QdrantVectorDB) Stats(ctx context.Context) (map[string]interface{}, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/collections/%s", q.baseURL, q.collection), nil)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qdrant stats failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return map[string]interface{}{
			"exists":       false,
			"points_count": 0,
		}, nil
	}

	var result struct {
		Result struct {
			PointsCount int `json:"points_count"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return map[string]interface{}{
		"exists":       true,
		"points_count": result.Result.PointsCount,
	}, nil
}

func (q *QdrantVectorDB) Collections(ctx context.Context) ([]string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/collections", q.baseURL), nil)
	resp, err := q.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qdrant collections failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			Collections []struct {
				Name string `json:"name"`
			} `json:"collections"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	var names []string
	for _, c := range result.Result.Collections {
		names = append(names, c.Name)
	}
	return names, nil
}

type ChromaVectorDB struct {
	baseURL    string
	collection string
	httpClient *http.Client
	apiKey     string
	batchSize  int
}

func NewChromaVectorDB(url, collection string, apiKey string, batchSize int) *ChromaVectorDB {
	return &ChromaVectorDB{
		baseURL:    url,
		collection: collection,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
		batchSize:  batchSize,
	}
}

func (c *ChromaVectorDB) Search(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]SearchResult, error) {
	payload := map[string]interface{}{
		"query_embeddings": [][]float32{embedding},
		"n_results":        topK,
		"include":          []string{"embeddings", "metadatas", "documents", "distances"},
	}

	if len(filters) > 0 {
		payload["where"] = filters
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, c.collection)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chroma search failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Documents [][]string                 `json:"documents"`
		Metadatas [][]map[string]interface{} `json:"metadatas"`
		Distances [][]float32                `json:"distances"`
	}

	json.Unmarshal(respBody, &result)

	var results []SearchResult
	if len(result.Documents) > 0 {
		for i, doc := range result.Documents[0] {
			distance := float32(0)
			if i < len(result.Distances[0]) {
				distance = 1 - result.Distances[0][i]
			}

			meta := map[string]interface{}{}
			if i < len(result.Metadatas[0]) {
				meta = result.Metadatas[0][i]
			}

			sr := SearchResult{
				Document: doc,
				Content:  doc,
				Score:    distance,
				Metadata: meta,
				Citation: Citation{
					Score: distance,
				},
			}

			results = append(results, sr)
		}
	}

	return results, nil
}

func (c *ChromaVectorDB) Store(ctx context.Context, doc Document) error {
	var ids []string
	var embeddings [][]float32
	var metadatas []map[string]interface{}
	var documents []string

	for _, chunk := range doc.Chunks {
		ids = append(ids, chunk.ID)
		embeddings = append(embeddings, chunk.Embedding)
		documents = append(documents, chunk.Content)

		meta := map[string]interface{}{
			"document":      doc.ID,
			"source":        doc.Source,
			"document_type": doc.DocumentType,
		}
		metadatas = append(metadatas, meta)
	}

	payload := map[string]interface{}{
		"ids":        ids,
		"embeddings": embeddings,
		"metadatas":  metadatas,
		"documents":  documents,
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, c.collection)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("chroma store failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("chroma store returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *ChromaVectorDB) Delete(ctx context.Context, docID string) error {
	payload := map[string]interface{}{
		"where": map[string]interface{}{
			"document": map[string]interface{}{
				"$eq": docID,
			},
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/v1/collections/%s/delete", c.baseURL, c.collection)
	req, _ := http.NewRequestWithContext(ctx, "DELETE", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("chroma delete failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("chroma delete returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *ChromaVectorDB) Health(ctx context.Context) (map[string]interface{}, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/heartbeat", c.baseURL), nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chroma health check failed: %w", err)
	}
	defer resp.Body.Close()

	return map[string]interface{}{"status": "healthy"}, nil
}

func (c *ChromaVectorDB) Stats(ctx context.Context) (map[string]interface{}, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/collections", c.baseURL), nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chroma stats failed: %w", err)
	}
	defer resp.Body.Close()

	return map[string]interface{}{"collections": 0}, nil
}

func (c *ChromaVectorDB) Collections(ctx context.Context) ([]string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/collections", c.baseURL), nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chroma collections failed: %w", err)
	}
	defer resp.Body.Close()

	var result []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	var names []string
	for _, c := range result {
		if name, ok := c["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names, nil
}
