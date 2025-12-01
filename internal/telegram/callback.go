package telegram

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/TheTeemka/telegram_bot_moodle_grades/pkg/utils"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *TelegramBot) HandleCallbacks(callback tapi.CallbackQuery) {
	fields := strings.Split(callback.Data, ":")
	switch fields[0] {
	case "crs":
		if len(fields) < 2 {
			slog.Warn("Invalid course callback data", "data", callback.Data)
			return
		}
		courseName := fields[1]
		b.CallbackCourse(courseName)
	default:
		slog.Warn("Unknown callback data", "data", callback.Data)
		return
	}
}

func (b *TelegramBot) CallbackCourse(compCourseFile string) {
	slog.Debug("Handling course callback", "course", compCourseFile)
	var courseFile string
	courseNames, err := b.gradeService.GetCourseNamesList()
	if err != nil {
		slog.Error("Failed to get course names list", "error", err)
		b.SendError("Failed to get course names list")
		return
	}

	for _, name := range courseNames {
		if utils.Compress(name) == compCourseFile {
			courseFile = name
			break
		}
	}

	if courseFile == "" {
		slog.Error("Course name not found for callback", "compCourseName", compCourseFile)
		b.SendError("Course not found")
		return
	}

	slog.Debug("Fetching grades for course", "course", courseFile)

	rows, err := b.gradeService.GetCourseFile(courseFile)
	if err != nil {
		slog.Error("Failed to get course grades file", "course", courseFile, "error", err)
		b.SendError("Failed to get course grades file for " + courseFile)
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Grades for course: %s (%d)\n\n", courseFile, len(rows))

	var messageRows []string
	for _, row := range rows {
		messageRows = append(messageRows, row.StringWithName())
	}

	sort.Strings(messageRows)

	for i, r := range messageRows {
		fmt.Fprintf(&sb, "%2d. %s\n", i+1, r)
	}

	message := sb.String()

	err = b.SendToTarget(message)
	if err != nil {
		slog.Error("Failed to send course grades", "error", err)
		b.SendError("Failed to send course grades for " + courseFile)
	}
}
