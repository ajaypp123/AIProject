package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

type Retriever struct {
	vectorDB  VectorDB
	embedding EmbeddingProvider
	config    RetrieverConfig
}

func NewRetriever(vectorDB VectorDB, embedding EmbeddingProvider, config RetrieverConfig) *Retriever {
	return &Retriever{
		vectorDB:  vectorDB,
		embedding: embedding,
		config:    config,
	}
}

func (r *Retriever) Retrieve(ctx context.Context, query string, filters map[string]interface{}) ([]SearchResult, error) {
	embedding, err := r.embedding.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	topK := r.config.TopK
	if topK == 0 {
		topK = 5
	}

	results, err := r.vectorDB.Search(ctx, embedding, topK*2, filters)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	switch r.config.Strategy {
	case "mmr":
		results = r.maximalMarginalRelevance(embedding, results, topK)
	case "similarity":
		fallthrough
	default:
		if len(results) > topK {
			results = results[:topK]
		}
	}

	for i := range results {
		if results[i].Score < r.config.ScoreThreshold {
			results = results[:i]
			break
		}
	}

	return results, nil
}

func (r *Retriever) maximalMarginalRelevance(queryEmb []float32, results []SearchResult, k int) []SearchResult {
	if len(results) <= k {
		return results
	}

	selected := make([]SearchResult, 0, k)
	remaining := make([]SearchResult, len(results))
	copy(remaining, results)

	if len(remaining) > 0 {
		selected = append(selected, remaining[0])
		remaining = append(remaining[:0], remaining[1:]...)
	}

	for len(selected) < k && len(remaining) > 0 {
		maxScore := float32(-1)
		maxIdx := 0

		for i, result := range remaining {
			relevance := result.Score
			maxMin := float32(1)

			for _, sel := range selected {
				sim := cosineSimilarity(getEmbedding(result), getEmbedding(sel))
				if sim < maxMin {
					maxMin = sim
				}
			}

			score := relevance - 0.5*maxMin
			if score > maxScore {
				maxScore = score
				maxIdx = i
			}
		}

		selected = append(selected, remaining[maxIdx])
		remaining = append(remaining[:maxIdx], remaining[maxIdx+1:]...)
	}

	return selected
}

func getEmbedding(result SearchResult) []float32 {
	if len(result.Citation.Source) > 0 {
		h := rand.Uint32()
		for range make([]byte, 384) {
			h = h*1103515245 + 12345
		}
		emb := make([]float32, 384)
		for i := range emb {
			emb[i] = float32(int(h)%1000) / 1000.0
		}
		return emb
	}
	return make([]float32, 384)
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

type PromptBuilder struct {
	systemPrompt string
}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		systemPrompt: `You are an expert Apache Spark knowledge assistant. Your role is to provide accurate, detailed, and helpful answers about Apache Spark, Hadoop, Delta Lake, Iceberg, and related big data technologies.

Guidelines:
1. Base your answers on the provided context documents
2. Be specific and technical in your responses
3. Include relevant code examples when appropriate
4. Cite your sources clearly
5. If information is not in the provided context, clearly state that
6. Provide practical, actionable advice
7. Explain complex concepts clearly
8. Keep answers concise but comprehensive`,
	}
}

func (pb *PromptBuilder) BuildChatPrompt(query string, context []SearchResult) string {
	prompt := fmt.Sprintf("%s\n\n", pb.systemPrompt)

	if len(context) > 0 {
		prompt += "Context from documentation:\n"
		prompt += "---\n"
		for i, result := range context {
			prompt += fmt.Sprintf("[Source %d] %s\n%s\n\n", i+1, result.Document, result.Content)
		}
		prompt += "---\n\n"
	}

	prompt += fmt.Sprintf("User Question: %s\n\nPlease provide a comprehensive answer based on the context provided above.", query)

	return prompt
}

func (pb *PromptBuilder) BuildSearchPrompt(query string) string {
	return fmt.Sprintf("Search Query: %s\n\nFind relevant documentation and examples.", query)
}

type CitationBuilder struct{}

func NewCitationBuilder() *CitationBuilder {
	return &CitationBuilder{}
}

func (cb *CitationBuilder) BuildCitations(results []SearchResult) []Citation {
	citations := make([]Citation, 0)

	for _, result := range results {
		citation := Citation{
			Document: result.Document,
			Source:   result.Citation.Source,
			Score:    result.Score,
		}

		if metadata, ok := result.Metadata["chunk_id"].(string); ok {
			citation.Section = metadata
		}

		citations = append(citations, citation)
	}

	sort.Slice(citations, func(i, j int) bool {
		return citations[i].Score > citations[j].Score
	})

	if len(citations) > 5 {
		citations = citations[:5]
	}

	return citations
}

func (cb *CitationBuilder) BuildRelatedDocuments(results []SearchResult) []Document {
	docMap := make(map[string]Document)

	for _, result := range results {
		if _, exists := docMap[result.Document]; !exists {
			docMap[result.Document] = Document{
				ID:      result.Document,
				Title:   result.Document,
				Source:  result.Citation.Source,
				Content: result.Content,
			}
		}
	}

	docs := make([]Document, 0, len(docMap))
	for _, doc := range docMap {
		docs = append(docs, doc)
	}

	return docs
}
