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

type LLMProvider interface {
	Generate(ctx context.Context, prompt string, history []Message) (string, error)
	GenerateStream(ctx context.Context, prompt string, history []Message, callback func(string)) error
}

type OllamaLLM struct {
	baseURL     string
	model       string
	maxTokens   int
	temperature float32
	client      *http.Client
}

func NewOllamaLLM(baseURL, model string, maxTokens int, temperature float32) *OllamaLLM {
	return &OllamaLLM{
		baseURL:     baseURL,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		client:      &http.Client{Timeout: 5 * time.Minute},
	}
}

func (o *OllamaLLM) Generate(ctx context.Context, prompt string, history []Message) (string, error) {
	fullPrompt := prompt
	for _, msg := range history {
		fullPrompt = fmt.Sprintf("%s\n%s: %s", fullPrompt, msg.Role, msg.Content)
	}

	payload := map[string]interface{}{
		"model":  o.model,
		"prompt": fullPrompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": o.temperature,
			"top_p":       0.9,
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/generate", o.baseURL), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama generation failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response, nil
}

func (o *OllamaLLM) GenerateStream(ctx context.Context, prompt string, history []Message, callback func(string)) error {
	fullPrompt := prompt
	for _, msg := range history {
		fullPrompt = fmt.Sprintf("%s\n%s: %s", fullPrompt, msg.Role, msg.Content)
	}

	payload := map[string]interface{}{
		"model":  o.model,
		"prompt": fullPrompt,
		"stream": true,
		"options": map[string]interface{}{
			"temperature": o.temperature,
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/generate", o.baseURL), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama stream failed: %w", err)
	}
	defer resp.Body.Close()

	scanner := bytes.NewBuffer(nil)
	io.Copy(scanner, resp.Body)

	for scanner.Len() > 0 {
		line, _ := scanner.ReadBytes('\n')
		if len(line) == 0 {
			break
		}

		var chunk struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		json.Unmarshal(line, &chunk)

		if chunk.Response != "" {
			callback(chunk.Response)
		}

		if chunk.Done {
			break
		}
	}

	return nil
}

type OpenAILLM struct {
	apiKey      string
	model       string
	maxTokens   int
	temperature float32
	client      *http.Client
}

func NewOpenAILLM(apiKey, model string, maxTokens int, temperature float32) *OpenAILLM {
	return &OpenAILLM{
		apiKey:      apiKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		client:      &http.Client{Timeout: 5 * time.Minute},
	}
}

func (o *OpenAILLM) Generate(ctx context.Context, prompt string, history []Message) (string, error) {
	messages := []map[string]string{
		{"role": "system", "content": "You are a helpful assistant for Apache Spark knowledge questions."},
	}

	for _, msg := range history {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	payload := map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  o.maxTokens,
		"temperature": o.temperature,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai generation failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return result.Choices[0].Message.Content, nil
}

func (o *OpenAILLM) GenerateStream(ctx context.Context, prompt string, history []Message, callback func(string)) error {
	messages := []map[string]string{
		{"role": "system", "content": "You are a helpful assistant for Apache Spark knowledge questions."},
	}

	for _, msg := range history {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	messages = append(messages, map[string]string{
		"role":    "user",
		"content": prompt,
	})

	payload := map[string]interface{}{
		"model":       o.model,
		"messages":    messages,
		"max_tokens":  o.maxTokens,
		"temperature": o.temperature,
		"stream":      true,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("openai stream failed: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
