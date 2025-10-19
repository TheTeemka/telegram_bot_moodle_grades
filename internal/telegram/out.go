package telegram

import (
	"log/slog"

	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *TelegramBot) Send(chatID int64, msg string) error {
	message := tapi.NewMessage(chatID, msg)
	message.ParseMode = tapi.ModeHTML
	_, err := b.bot.Send(message)
	return err
}

func (b *TelegramBot) SendToTarget(msg string) error {
	return b.Send(b.targetID, msg)
}

func (b *TelegramBot) StartMessage() {
	err := b.SendToTarget("✳️ Bot has started!")
	if err != nil {
		slog.Error("Failed to send start message", "error", err)
	}
}

func (b *TelegramBot) DeadMessage() {
	err := b.SendToTarget("☠️ Bot shutting down")
	if err != nil {
		slog.Error("Failed to send shutdown message", "error", err)
	}
}
