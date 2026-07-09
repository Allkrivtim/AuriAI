package telegram

import (
	"AssistantAI/internal/adapters/telegram/handlers"
	"log"
	"os"
	"strconv"

	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func InitTgBot() {
	bot, err := tb.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {

		bot.Debug = false
	}
	log.Printf("Authorized on bot @%s", bot.Self.UserName)

	updateConfig := tb.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message.IsCommand() {
			continue
		}
		if update.Message != nil {
			go handlers.HandleMessage(bot, &update)
		}
	}

	return
}
