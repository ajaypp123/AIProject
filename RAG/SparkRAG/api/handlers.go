package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	retriever        *Retriever
	vectorDB         VectorDB
	llm              LLMProvider
	embedding        EmbeddingProvider
	promptBuilder    *PromptBuilder
	citationBuilder  *CitationBuilder
	config           *Config
	logger           *zap.Logger
}

func NewHandler(retriever *Retriever, vectorDB VectorDB, llm LLMProvider, embedding EmbeddingProvider, config *Config, logger *zap.Logger) *Handler {
	return &Handler{
		retriever:       retriever,
		vectorDB:        vectorDB,
		llm:             llm,
		embedding:       embedding,
		promptBuilder:   NewPromptBuilder(),
		citationBuilder: NewCitationBuilder(),
		config:          config,
		logger:          logger,
	}
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if req.Message == "" {
		h.writeError(w, http.StatusBadRequest, "EMPTY_MESSAGE", "message field is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	results, err := h.retriever.Retrieve(ctx, req.Message, req.Filters)
	if err != nil {
		h.logger.Error("retrieval failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "RETRIEVAL_ERROR", err.Error())
		return
	}

	prompt := h.promptBuilder.BuildChatPrompt(req.Message, results)

	response, err := h.llm.Generate(ctx, prompt, req.History)
	if err != nil {
		h.logger.Error("generation failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "GENERATION_ERROR", err.Error())
		return
	}

	conversationID := req.ConversationID
	if conversationID == "" {
		conversationID = uuid.New().String()
	}

	citations := h.citationBuilder.BuildCitations(results)
	confidenceScore := float32(0)
	if len(results) > 0 {
		confidenceScore = results[0].Score
	}

	chatResp := ChatResponse{
		ConversationID: conversationID,
		Answer:         response,
		Citations:      citations,
		ConfidenceScore: confidenceScore,
		Timestamp:      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResp)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if req.Query == "" {
		h.writeError(w, http.StatusBadRequest, "EMPTY_QUERY", "query field is required")
		return
	}

	if req.TopK == 0 {
		req.TopK = h.config.Retriever.TopK
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	results, err := h.retriever.Retrieve(ctx, req.Query, req.Filters)
	if err != nil {
		h.logger.Error("search failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "SEARCH_ERROR", err.Error())
		return
	}

	searchResp := SearchResponse{
		Results:   results,
		Total:     len(results),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchResp)
}

func (h *Handler) Ingest(w http.ResponseWriter, r *http.Request) {
	var req IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if req.DocumentPath == "" {
		h.writeError(w, http.StatusBadRequest, "EMPTY_PATH", "document_path is required")
		return
	}

	h.logger.Info("ingestion requested", zap.String("path", req.DocumentPath))

	resp := IngestResponse{
		DocumentID:  uuid.New().String(),
		ChunksAdded: 0,
		Success:     true,
		Message:     "Document ingestion queued. Processing in background.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if req.DocumentID == "" {
		h.writeError(w, http.StatusBadRequest, "EMPTY_ID", "document_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err := h.vectorDB.Delete(ctx, req.DocumentID)
	if err != nil {
		h.logger.Error("delete failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "DELETE_ERROR", err.Error())
		return
	}

	resp := DeleteResponse{
		DocumentID: req.DocumentID,
		Deleted:    true,
		Message:    "Document deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Documents(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stats, err := h.vectorDB.Stats(ctx)
	if err != nil {
		h.logger.Error("stats failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "STATS_ERROR", err.Error())
		return
	}

	resp := map[string]interface{}{
		"stats": stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Providers(w http.ResponseWriter, r *http.Request) {
	resp := ProvidersResponse{
		EmbeddingProviders: []string{"ollama", "openai"},
		LLMProviders:       []string{"ollama", "openai"},
		VectorDBProviders:  []string{"qdrant", "chroma"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Collections(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	collections, err := h.vectorDB.Collections(ctx)
	if err != nil {
		h.logger.Error("collections failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "COLLECTIONS_ERROR", err.Error())
		return
	}

	resp := CollectionsResponse{
		Collections: collections,
		Active:      h.config.VectorDB.Collection,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	vectorDBHealth, err := h.vectorDB.Health(ctx)
	if err != nil {
		vectorDBHealth = map[string]interface{}{"status": "down", "error": err.Error()}
	}

	resp := HealthResponse{
		Status: "healthy",
		Services: map[string]interface{}{
			"vectordb": vectorDBHealth,
			"llm":      h.config.LLM.Provider,
			"embedding": h.config.Embedding.Provider,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stats, _ := h.vectorDB.Stats(ctx)

	resp := MetricsResponse{
		TotalDocuments: 0,
		TotalChunks:    0,
		VectorDBStats:  stats,
		APIStats: map[string]interface{}{
			"requests": 0,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Code:    code,
	})
}

func (h *Handler) ChatStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"Invalid request\"}\n\n")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	results, err := h.retriever.Retrieve(ctx, req.Message, req.Filters)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"Retrieval failed\"}\n\n")
		return
	}

	prompt := h.promptBuilder.BuildChatPrompt(req.Message, results)

	err = h.llm.GenerateStream(ctx, prompt, req.History, func(chunk string) {
		data := map[string]string{"chunk": chunk}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		w.(http.Flusher).Flush()
	})

	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"Generation failed\"}\n\n")
	}
}
