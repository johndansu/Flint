package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	apiURL          = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
	maxTokens       = 4096
)

// Client calls the Anthropic Messages API directly over HTTP.
type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

// New creates a Client. model defaults to claude-sonnet-4-20250514 if empty.
func New(apiKey, model string) *Client {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 120 * time.Second},
	}
}

// Message is one turn in a conversation.
type Message struct {
	Role    string `json:"role"`    // "user" | "assistant"
	Content string `json:"content"`
}

// Complete sends messages and returns the full response text (non-streaming).
func (c *Client) Complete(ctx context.Context, system string, messages []Message) (string, error) {
	body, err := c.buildRequest(system, messages, false)
	if err != nil {
		return "", err
	}

	req, err := c.newRequest(ctx, body)
	if err != nil {
		return "", err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ai: HTTP %d: %s", resp.StatusCode, raw)
	}

	var out struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("ai: decode response: %w", err)
	}
	if len(out.Content) == 0 {
		return "", fmt.Errorf("ai: empty response")
	}
	return out.Content[0].Text, nil
}

// Stream sends messages and streams response tokens to the provided callback.
// The callback receives each text delta as it arrives. Returns full text when done.
func (c *Client) Stream(ctx context.Context, system string, messages []Message, delta func(string)) (string, error) {
	body, err := c.buildRequest(system, messages, true)
	if err != nil {
		return "", err
	}

	req, err := c.newRequest(ctx, body)
	if err != nil {
		return "", err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai: stream request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ai: HTTP %d: %s", resp.StatusCode, raw)
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var event struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			sb.WriteString(event.Delta.Text)
			if delta != nil {
				delta(event.Delta.Text)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return sb.String(), fmt.Errorf("ai: stream read: %w", err)
	}
	return sb.String(), nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (c *Client) buildRequest(system string, messages []Message, stream bool) ([]byte, error) {
	payload := map[string]any{
		"model":      c.model,
		"max_tokens": maxTokens,
		"messages":   messages,
		"stream":     stream,
	}
	if system != "" {
		payload["system"] = system
	}
	return json.Marshal(payload)
}

func (c *Client) newRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	return req, nil
}
