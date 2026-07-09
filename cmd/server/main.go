package main

import (
	"AssistantAI/internal/adapters/telegram"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("No .env file")
	}

	telegram.InitTgBot()
}
