package websearch

import (
	"AuriAI/internal/core"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (t *Tool) Spec() core.ToolSpec {
	return core.ToolSpec{
		Name:        "web_search",
		Description: "Search the web for current information. Use when you need up-to-date facts, news, or anything you don't know.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The search query",
				},
			},
			"required": []string{"query"},
		},
	}
}

func (t *Tool) Invoke(ctx context.Context, args string) (string, error) {
	var input struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return "", err
	}
	response, err := t.search(ctx, TavilySearchRequest{
		Query:         input.Query,
		MaxResults:    5,
		IncludeAnswer: true,
	})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	if response.Answer != "" {
		sb.WriteString("Answer: " + response.Answer + "\n\n")
	}
	for i, r := range response.Results {
		fmt.Fprintf(&sb, "%d. %s\n%s\n%s\n\n", i+1, r.Title, r.URL, r.Content)
	}
	if sb.Len() == 0 {
		return "No results found.", nil
	}
	return sb.String(), nil
}

func (t *Tool) search(ctx context.Context, searchRequest TavilySearchRequest) (TavilySearchResponse, error) {
	body, err := json.Marshal(searchRequest)
	if err != nil {
		return TavilySearchResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx,
		"POST",
		"https://api.tavily.com/search",
		bytes.NewReader(body),
	)
	if err != nil {
		return TavilySearchResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	client := t.client

	resp, err := client.Do(req)
	if err != nil {
		return TavilySearchResponse{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return TavilySearchResponse{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TavilySearchResponse{}, fmt.Errorf("tavily error: status=%d body=%s", resp.StatusCode, string(data))
	}

	var result TavilySearchResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return TavilySearchResponse{}, err
	}

	return result, nil
}
