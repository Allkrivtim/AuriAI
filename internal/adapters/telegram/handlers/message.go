package handlers

import (
	"AssistantAI/internal/core"
	"context"
	"log"

	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessage(bot *tb.BotAPI, update *tb.Update, engine core.Engine) error {
	log.Printf("MH|[%s] %s", update.Message.From.UserName, update.Message.Text)
	ctx := context.Background()
	inmsg, err := engine.Handle(ctx, core.InboundMessage{"1", "TG", update.Message.Text})
	if err != nil {
		return err
	}

	msg := tb.NewMessage(update.Message.Chat.ID, inmsg.Text)

	msg.ReplyToMessageID = update.Message.MessageID

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
	return err
}
