package main

import (
	"AuriAI/internal/adapters/llm/openai"
	"AuriAI/internal/adapters/store/sqlite"
	"AuriAI/internal/adapters/telegram"
	"AuriAI/internal/adapters/tools/memory"
	"AuriAI/internal/adapters/tools/notes"
	"AuriAI/internal/adapters/tools/websearch"
	"AuriAI/internal/core"
	"os"
)

func main() {
	//if err := godotenv.Load(); err != nil {
	//	panic("No .env file")
	//}
	llm := openai.NewClient(os.Getenv("LLM_API_KEY"), os.Getenv("LLM_MODEL"), os.Getenv("LLM_URL"))
	store, err := sqlite.NewStore("storage/core.sqlite")
	if err != nil {
		panic(err)
	}
	b, err := os.ReadFile("prompts/base.md")
	if err != nil {
		panic(err)
	}
	basePrompt := string(b)

	toolRegistry := core.NewRegistry()
	toolRegistry.Register(websearch.NewTool(os.Getenv("SEARCH_PROVIDER_API_KEY")))
	toolRegistry.Register(memory.NewTool(store))
	toolRegistry.Register(notes.NewTool(store))

	engine := core.NewEngine(llm, store, store, basePrompt, toolRegistry)

	telegram.InitTgBot(engine, os.Getenv("TG_BOT_TOKEN"))
}
