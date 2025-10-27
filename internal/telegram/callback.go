package telegram

import (
	"log/slog"
	"strings"

	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *TelegramBot) HandleCallbacks(callback tapi.CallbackQuery) {
	fields := strings.Split(callback.Data, ":")
	switch fields[0] {
	case "course":
	default:
		slog.Warn("Unknown callback data", "data", callback.Data)
		return
	}
}

func (b *TelegramBot) CallbackCourse(courseName string) {
	rows, err := b.gradeService.GetCourseGrades(courseName)
	if err != nil {
		slog.Error("Failed to get course grades", "course", courseName, "error", err)
		b.SendError("Failed to get course grades for " + courseName)
		return
	}

	message := "Grades for course: " + courseName + "\n"
	for _, row := range rows {
		message += row.StringWithName() + "\n"
	}

	err = b.SendToTarget(message)
	if err != nil {
		slog.Error("Failed to send course grades", "error", err)
		b.SendError("Failed to send course grades for " + courseName)
	}
}
