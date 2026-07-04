package main

import "time"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Citation struct {
	Source      string `json:"source"`
	Document    string `json:"document"`
	Section     string `json:"section,omitempty"`
	Page        int    `json:"page,omitempty"`
	Line        int    `json:"line,omitempty"`
	URL         string `json:"url,omitempty"`
	Score       float32 `json:"score,omitempty"`
}

type ChatRequest struct {
	Message      string            `json:"message"`
	ConversationID string           `json:"conversation_id,omitempty"`
	History      []Message         `json:"history,omitempty"`
	Filters      map[string]string `json:"filters,omitempty"`
	TopK         int               `json:"top_k,omitempty"`
	Strategy     string            `json:"strategy,omitempty"`
}

type ChatResponse struct {
	ConversationID string      `json:"conversation_id"`
	Answer         string      `json:"answer"`
	Citations      []Citation  `json:"citations"`
	RelatedDocs    []Document  `json:"related_docs,omitempty"`
	SuggestedQuestions []string `json:"suggested_questions,omitempty"`
	ConfidenceScore float32     `json:"confidence_score"`
	Timestamp      time.Time   `json:"timestamp"`
}

type SearchRequest struct {
	Query      string            `json:"query"`
	Filters    map[string]string `json:"filters,omitempty"`
	TopK       int               `json:"top_k,omitempty"`
	Strategy   string            `json:"strategy,omitempty"`
}

type SearchResult struct {
	Document string            `json:"document"`
	Content  string            `json:"content"`
	Score    float32           `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
	Citation Citation          `json:"citation"`
}

type SearchResponse struct {
	Results   []SearchResult `json:"results"`
	Total     int            `json:"total"`
	Timestamp time.Time      `json:"timestamp"`
}

type Document struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	Source     string                 `json:"source"`
	DocumentType string                `json:"document_type"`
	Version    string                 `json:"version,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Language   string                 `json:"language,omitempty"`
	LastModified time.Time             `json:"last_modified"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Chunks     []Chunk                `json:"chunks,omitempty"`
}

type Chunk struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Embedding []float32             `json:"embedding,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Score    float32                `json:"score,omitempty"`
}

type IngestRequest struct {
	DocumentPath string            `json:"document_path"`
	DocumentType string            `json:"document_type,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type IngestResponse struct {
	DocumentID string `json:"document_id"`
	ChunksAdded int    `json:"chunks_added"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

type DeleteRequest struct {
	DocumentID string `json:"document_id"`
}

type DeleteResponse struct {
	DocumentID string `json:"document_id"`
	Deleted    bool   `json:"deleted"`
	Message    string `json:"message"`
}

type HealthResponse struct {
	Status     string                 `json:"status"`
	Services   map[string]interface{} `json:"services"`
	Timestamp  time.Time              `json:"timestamp"`
}

type MetricsResponse struct {
	TotalDocuments  int                    `json:"total_documents"`
	TotalChunks     int                    `json:"total_chunks"`
	VectorDBStats   map[string]interface{} `json:"vectordb_stats"`
	APIStats        map[string]interface{} `json:"api_stats"`
	Timestamp       time.Time              `json:"timestamp"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

type ProvidersResponse struct {
	EmbeddingProviders []string `json:"embedding_providers"`
	LLMProviders       []string `json:"llm_providers"`
	VectorDBProviders  []string `json:"vectordb_providers"`
}

type CollectionsResponse struct {
	Collections []string `json:"collections"`
	Active      string   `json:"active"`
}
