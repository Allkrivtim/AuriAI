package websearch

import (
	"net/http"
	"time"
)

type Tool struct {
	apiKey string
	client *http.Client
}

func NewTool(apiKey string) *Tool {
	return &Tool{
		apiKey: apiKey,
		client: &http.Client{Timeout: 15 * time.Second}, // таймаут обязателен!
	}
}

type TavilySearchRequest struct {
	Query             string   `json:"query"`
	SearchDepth       string   `json:"search_depth,omitempty"` // "basic" или "advanced"
	Topic             string   `json:"topic,omitempty"`        // "general", "news", etc.
	MaxResults        int      `json:"max_results,omitempty"`
	IncludeAnswer     bool     `json:"include_answer,omitempty"`
	IncludeRawContent bool     `json:"include_raw_content,omitempty"`
	IncludeDomains    []string `json:"include_domains,omitempty"`
	ExcludeDomains    []string `json:"exclude_domains,omitempty"`
}
type TavilySearchResponse struct {
	Query   string `json:"query"`
	Answer  string `json:"answer"`
	Results []struct {
		Title      string  `json:"title"`
		URL        string  `json:"url"`
		Content    string  `json:"content"`
		RawContent string  `json:"raw_content"`
		Score      float64 `json:"score"`
	} `json:"results"`
}
