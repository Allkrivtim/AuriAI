package main

import (
	"AssistantAI/internal/adapters/llm/openrouter"
	"AssistantAI/internal/adapters/store/sqlite"
	"AssistantAI/internal/adapters/telegram"
	"AssistantAI/internal/core"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file")
	}
	llm := openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"), os.Getenv("LLM_MODEL"))
	store, err := sqlite.NewStore("")
	if err != nil {
		panic(err)
	}
	engine := core.NewEngine(llm, store)

	telegram.InitTgBot(engine, os.Getenv("TG_BOT_TOKEN"))
}
