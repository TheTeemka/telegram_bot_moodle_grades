package telegram

import (
	"fmt"
	"log/slog"

	"github.com/TheTeemka/telegram_bot_moodle_grades/pkg/utils"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *TelegramBot) HandleCommands(update tapi.Update) {
	if update.Message != nil {
		switch update.Message.Command() {
		case "start":
			b.HandleStart()
		case "sync":
			b.HandleManualSync()
		case "status":
			b.HandlerStatus()
		case "list":
			b.HandleList()
		}
	}
}

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

func (b *TelegramBot) HandleList() {
	courseNames, err := b.gradeService.GetCourseNamesList()
	if err != nil {
		slog.Error("Failed to get course names", "error", err)
		b.SendError("Failed to get course names")
		return
	}

	slog.Debug("Sending course list", "courses", courseNames)
	var keyboard [][]tapi.InlineKeyboardButton
	for _, courseName := range courseNames {
		keyboard = append(keyboard, []tapi.InlineKeyboardButton{
			tapi.NewInlineKeyboardButtonData(courseName, "crs:"+utils.Compress(courseName)),
		})
	}

	err = b.SendToTargetWithKeyboard("Available courses:", keyboard)
	if err != nil {
		slog.Error("Failed to send course list", "error", err)
		b.SendError("Failed to send course list")
	}
}
