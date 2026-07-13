package main

import (
	"AuriAI/internal/adapters/llm/openai"
	"AuriAI/internal/adapters/store/sqlite"
	"AuriAI/internal/adapters/telegram"
	"AuriAI/internal/core"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file")
	}
	llm := openai.NewClient(os.Getenv("LLM_API_KEY"), os.Getenv("LLM_MODEL"), os.Getenv("LLM_URL"))
	store, err := sqlite.NewStore("storage/assistant.sqlite")
	if err != nil {
		panic(err)
	}
	b, err := os.ReadFile("internal/prompts/base.md")
	if err != nil {
		panic(err)
	}
	basePrompt := string(b)

	engine := core.NewEngine(llm, store, basePrompt)

	telegram.InitTgBot(engine, os.Getenv("TG_BOT_TOKEN"))
}
