package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"quokka-ai-bot/config"
	"time"
)

type DeepSeekRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type DeepSeekClient struct {
	APIKey     string
	HTTPClinet *http.Client
	BaseURL    string
}

func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{
		APIKey: apiKey,
		HTTPClinet: &http.Client{
			Timeout: 2 * time.Minute,
		},
		BaseURL: config.Load().BaseURL,
	}
}

func (c *DeepSeekClient) ChatCompeletion(ctx context.Context, req DeepSeekRequest) (string, error) {
	reqBody, err := json.Marshal(req) // Marshal the request to json to send the request to deepseek api
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody)) // Creating a request to api
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	httpReq.Header.Set("Content-type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTPClinet.Do(httpReq) // We execute the request
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(body))
	}
	var response DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil { // Decode the response from the api
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(response.Choices) == 0 { // If we haven't received a response
		if response.Error.Message != "" {
			return "", fmt.Errorf("api error: %s", response.Error.Message)
		}
		return "", fmt.Errorf("no choices in response")
	}
	return response.Choices[0].Message.Content, nil // If we receive a response - return it
}
