package telegram

import (
	"AuriAI/internal/adapters/telegram/handlers"
	"AuriAI/internal/core"
	"log"

	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func InitTgBot(engine core.Engine, token string) {
	bot, err := tb.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on bot @%s", bot.Self.UserName)

	updateConfig := tb.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.IsCommand() {
			continue
		}
		if update.Message != nil {
			go handlers.HandleMessage(bot, &update, engine)
		}
	}

	return
}
