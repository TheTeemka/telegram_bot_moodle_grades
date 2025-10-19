package telegram

import (
	"encoding/json"
	"fmt"
	"log/slog"

	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *TelegramBot) HandleStart() {
	err := b.SendToTarget("Bot is running!")
	if err != nil {
		slog.Error("Failed to send message", "error", err)
	}
}

func (b *TelegramBot) HandleSync() error {
	changes, err := b.gradeService.ParseAndCompare()
	if err != nil {
		b.SendToTarget(err.Error())
		slog.Error("Failed to parse and compare", "error", err)
		return err
	}

	for _, change := range changes {
		err := b.SendToTarget(change.ToHTMLString())
		if err != nil {
			slog.Error("Failed to send change message", "error", err)
		}
	}
	return nil
}

func (b *TelegramBot) HandleManualSync() error {
	err := b.SendToTarget("Manual sync triggered")
	if err != nil {
		slog.Error("Failed to send sync message", "error", err)
		return err
	}

	err = b.HandleSync()
	if err != nil {
		slog.Error("Manual sync failed", "error", err)
		return err
	}

	err = b.SendToTarget("Manual sync finished")
	if err != nil {
		slog.Error("Failed to send sync message", "error", err)
		return err
	}
	return nil
}

func (b *TelegramBot) HandlerStatus() {
	msg := fmt.Sprintf("Last parsed at: %s", b.gradeService.LastTimeParsed.Format("2006-01-02 15:04:05"))
	err := b.SendToTarget(msg)
	if err != nil {
		slog.Error("Failed to send status", "error", err)
	}
}

func (b *TelegramBot) IsFromMe(update tapi.Update) bool {
	if update.Message != nil {
		if update.Message.Chat.ID != b.targetID {
			s, err := json.Marshal(update.Message.Chat)
			if err != nil {
				slog.Error("Failed to marshal chat", "error", err)
			}
			slog.Warn("Received message from unauthorized user",
				"user", string(s),
				"msg_content", update.Message.Text)

			err = b.Send(update.Message.Chat.ID,
				"❗️ кто вы такие, я вас не звал. Идете <tg-spoiler> домой пожалуйста 😊 </tg-spoiler>.")
			if err != nil {
				slog.Error("Failed to send unauthorized message", "error", err)
			}
			return false
		}
	}
	return true
}
