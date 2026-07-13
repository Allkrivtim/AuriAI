package openai

import "github.com/openai/openai-go"

type Client struct {
	api   openai.Client
	model string
	url   string
}
