package handlers

import (
	"AuriAI/internal/core"
	"context"
	"fmt"
	"log"
	"time"

	tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// telegramMessageLimit is Telegram's hard cap on a single message's text length.
const telegramMessageLimit = 4096

const helpText = "Просто напишите мне сообщение, и я отвечу.\n\n" +
	"Команды:\n" +
	"/start — начать\n" +
	"/help — эта справка"

// HandleCommand handles bot commands (/start, /help, ...). Unknown commands
// are ignored so the bot doesn't spam replies to commands meant for other bots.
func HandleCommand(bot *tb.BotAPI, msg *tb.Message) {
	var text string
	switch msg.Command() {
	case "start":
		text = "Привет! Я на связи. " + helpText
	case "help":
		text = helpText
	default:
		return
	}

	reply := tb.NewMessage(msg.Chat.ID, text)
	if _, err := bot.Send(reply); err != nil {
		log.Printf("[telegram] error sending command reply to chat %d: %v", msg.Chat.ID, err)
	}
}

// HandleMessage sends the message text to the engine and replies with the
// result. Each Telegram chat maps to its own conversation session, so
// different chats (and different users in private chats) never share history.
func HandleMessage(bot *tb.BotAPI, msg *tb.Message, engine core.Engine) {
	sessionID := fmt.Sprintf("tg:%d", msg.Chat.ID)

	from := "unknown"
	if msg.From != nil {
		from = msg.From.UserName
	}
	log.Printf("[telegram] <- session=%s from=%s text=%q", sessionID, from, msg.Text)

	typing := tb.NewChatAction(msg.Chat.ID, tb.ChatTyping)
	if _, err := bot.Request(typing); err != nil {
		log.Printf("[telegram] error sending typing action to chat %d: %v", msg.Chat.ID, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	out, err := engine.Handle(ctx, core.InboundMessage{SessionID: sessionID, Provider: "telegram", Text: msg.Text})
	if err != nil {
		log.Printf("[telegram] engine error for session %s: %v", sessionID, err)
		reply := tb.NewMessage(msg.Chat.ID, "Извините, произошла ошибка при обработке сообщения. Попробуйте ещё раз.")
		reply.ReplyToMessageID = msg.MessageID
		if _, sendErr := bot.Send(reply); sendErr != nil {
			log.Printf("[telegram] error sending error notice to chat %d: %v", msg.Chat.ID, sendErr)
		}
		return
	}

	log.Printf("[telegram] -> session=%s text=%q", sessionID, out.Text)

	for i, chunk := range splitMessage(out.Text, telegramMessageLimit) {
		reply := tb.NewMessage(msg.Chat.ID, chunk)
		if i == 0 {
			reply.ReplyToMessageID = msg.MessageID
		}
		if _, err := bot.Send(reply); err != nil {
			log.Printf("[telegram] error sending message to chat %d: %v", msg.Chat.ID, err)
			return
		}
	}
}

// splitMessage breaks text into chunks no longer than limit, preferring to
// break on newlines/spaces so words aren't cut in half. Telegram rejects
// messages over 4096 characters outright, so long LLM replies must be split.
func splitMessage(text string, limit int) []string {
	if text == "" {
		return []string{""}
	}

	runes := []rune(text)
	if len(runes) <= limit {
		return []string{text}
	}

	var chunks []string
	for len(runes) > 0 {
		if len(runes) <= limit {
			chunks = append(chunks, string(runes))
			break
		}

		cut := limit
		if idx := lastBreak(runes[:limit]); idx > 0 {
			cut = idx
		}

		chunks = append(chunks, string(runes[:cut]))
		runes = runes[cut:]
	}
	return chunks
}

// lastBreak returns the index just after the last newline or space in runes,
// or -1 if none was found.
func lastBreak(runes []rune) int {
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == '\n' || runes[i] == ' ' {
			return i + 1
		}
	}
	return -1
}
