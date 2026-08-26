package draft

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	defaultAnthropicURL = "https://api.anthropic.com"
	defaultOpenAIURL    = "https://api.openai.com"
	modelAnthropic      = "claude-sonnet-4-5"
	modelOpenAI         = "gpt-5-mini"
)

// readJSONField is a test helper pulling one top-level field from a request body.
func readJSONField(t interface {
	Fatalf(string, ...interface{})
}, r *http.Request, field string,
) string {
	var m map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	v, _ := m[field].(string)
	return v
}

type anthropicProvider struct {
	key, base string
}

// NewAnthropic drafts via the Anthropic Messages API.
func NewAnthropic(key, baseURL string) Provider { return anthropicProvider{key: key, base: baseURL} }

func (a anthropicProvider) Draft(ctx context.Context, p PRData) (string, error) {
	prompt, err := BuildPrompt(p)
	if err != nil {
		return "", err
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model":      modelAnthropic,
		"max_tokens": 4000,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	req, rerr := http.NewRequestWithContext(ctx, http.MethodPost, a.base+"/v1/messages", bytes.NewReader(body))
	if rerr != nil {
		return "", rerr
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", a.key)
	req.Header.Set("anthropic-version", "2023-06-01")
	text, status, herr := doJSON(req)
	if herr != nil {
		return "", fmt.Errorf("anthropic %d: %w", status, herr)
	}
	var resp struct {
		Content []struct{ Text string } `json:"content"`
	}
	if jerr := json.Unmarshal(text, &resp); jerr != nil || len(resp.Content) == 0 {
		return "", fmt.Errorf("anthropic %d: unexpected response shape (%d bytes)", status, len(text))
	}
	return resp.Content[0].Text, nil
}

type openAIProvider struct {
	key, base string
}

// NewOpenAI drafts via the OpenAI chat completions API.
func NewOpenAI(key, baseURL string) Provider { return openAIProvider{key: key, base: baseURL} }

func (o openAIProvider) Draft(ctx context.Context, p PRData) (string, error) {
	prompt, err := BuildPrompt(p)
	if err != nil {
		return "", err
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model": modelOpenAI,
		"max_completion_tokens": 4000,
		"messages":              []map[string]string{{"role": "user", "content": prompt}},
	})
	req, rerr := http.NewRequestWithContext(ctx, http.MethodPost, o.base+"/v1/chat/completions", bytes.NewReader(body))
	if rerr != nil {
		return "", rerr
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+o.key)
	text, status, herr := doJSON(req)
	if herr != nil {
		return "", fmt.Errorf("openai %d: %w", status, herr)
	}
	var resp struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
	}
	if jerr := json.Unmarshal(text, &resp); jerr != nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai %d: unexpected response shape (%d bytes)", status, len(text))
	}
	return resp.Choices[0].Message.Content, nil
}

func doJSON(req *http.Request) ([]byte, int, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode >= 300 {
		snippet := string(b)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		return b, resp.StatusCode, fmt.Errorf("%s", snippet)
	}
	return b, resp.StatusCode, nil
}
