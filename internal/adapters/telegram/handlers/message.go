package handlers

import (
	"AssistantAI/internal/core"
	"log"

	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessage(bot *tb.BotAPI, update *tb.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	msg_text, err := core.AskAI(update.Message.Text)
	if err != nil {
		log.Println(err)
	}

	msg := tb.NewMessage(update.Message.Chat.ID, msg_text)

	msg.ReplyToMessageID = update.Message.MessageID

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
