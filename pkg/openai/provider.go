package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type Provider struct {
	apiKey     string
	httpClient *http.Client
	logger     *slog.Logger
}

type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint"`
	ServiceTier       string   `json:"service_tier"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Logprobs     *any    `json:"logprobs"`
}

type Message struct {
	Role        string  `json:"role"`
	Content     string  `json:"content"`
	Refusal     *string `json:"refusal"`
	Annotations []any   `json:"annotations"`
}

type Usage struct {
	PromptTokens            int                     `json:"prompt_tokens"`
	CompletionTokens        int                     `json:"completion_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	PromptTokensDetails     PromptTokensDetails     `json:"prompt_tokens_details"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
}

type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
	AudioTokens  int `json:"audio_tokens"`
}

type CompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens"`
	AudioTokens              int `json:"audio_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

func NewProvider(apiKey string, logger *slog.Logger) *Provider {
	return &Provider{
		apiKey:     apiKey,
		logger:     logger,
		httpClient: http.DefaultClient,
	}
}

func (o *Provider) FetchAiResponse(contentToSend string) (string, error) {
	body := map[string]any{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": contentToSend},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		o.logger.Error("Failed to marshal OpenAI request body", "error", err)
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		o.logger.Error("Failed to create OpenAI HTTP request", "error", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		o.logger.Error("OpenAI API request failed", "error", err)
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errorBody); err == nil {
			o.logger.Error("OpenAI API returned error",
				"status_code", resp.StatusCode,
				"status", resp.Status,
				"error", errorBody,
			)
			return "", fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, resp.Status)
		}

		o.logger.Error("OpenAI API returned error",
			"status_code", resp.StatusCode,
			"status", resp.Status,
		)
		return "", fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		o.logger.Error("Failed to decode OpenAI response", "error", err)
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	o.logger.Debug("OpenAI Response",
		"id", result.ID,
		"model", result.Model,
		"completion_tokens", result.Usage.CompletionTokens,
		"prompt_tokens", result.Usage.PromptTokens,
	)

	if len(result.Choices) == 0 {
		o.logger.Error("OpenAI returned empty choices array")
		return "", fmt.Errorf("no choices returned from OpenAI")
	}

	content := result.Choices[0].Message.Content
	o.logger.Debug("OpenAI Content", "length", len(content))

	return content, nil
}
