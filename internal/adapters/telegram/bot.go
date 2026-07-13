package telegram

import (
	"AuriAI/internal/adapters/telegram/handlers"
	"AuriAI/internal/core"
	"log"
	"os"
	"os/signal"
	"syscall"

	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// InitTgBot starts long-polling Telegram for updates and dispatches each
// message to the engine. It blocks until an interrupt/terminate signal is
// received or the updates channel is closed.
func InitTgBot(engine core.Engine, token string) {
	bot, err := tb.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("[telegram] authorized as @%s", bot.Self.UserName)

	updateConfig := tb.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := bot.GetUpdatesChan(updateConfig)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				return
			}
			handleUpdate(bot, update, engine)

		case <-stop:
			log.Println("[telegram] shutdown signal received, stopping...")
			bot.StopReceivingUpdates()
			return
		}
	}
}

func handleUpdate(bot *tb.BotAPI, update tb.Update, engine core.Engine) {
	msg := update.Message
	if msg == nil {
		return
	}

	if msg.IsCommand() {
		handlers.HandleCommand(bot, msg)
		return
	}

	if !shouldRespond(bot, msg) {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[telegram] recovered panic while handling chat %d: %v", msg.Chat.ID, r)
			}
		}()
		handlers.HandleMessage(bot, msg, engine)
	}()
}

// shouldRespond decides whether the bot should reply to a non-command
// message. Private chats always get a reply; in groups/supergroups the bot
// only replies when it is directly mentioned or the message is a reply to
// one of its own messages, so it doesn't talk over every group message.
func shouldRespond(bot *tb.BotAPI, msg *tb.Message) bool {
	if msg.Chat.IsPrivate() {
		return true
	}

	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From != nil && msg.ReplyToMessage.From.ID == bot.Self.ID {
		return true
	}

	for _, e := range msg.Entities {
		if e.Type != "mention" {
			continue
		}
		if e.Offset+e.Length > len(msg.Text) {
			continue
		}
		if msg.Text[e.Offset:e.Offset+e.Length] == "@"+bot.Self.UserName {
			return true
		}
	}

	return false
}
