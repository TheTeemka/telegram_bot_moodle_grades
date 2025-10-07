package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/config"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/service"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/telegram"
)

func main() {
	cfg := config.Load()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	fetcher := service.NewMoodleFetcher(cfg.MoodleConfig)
	gradeService := service.NewGradeService(fetcher)

	slog.Info("Starting initial parse and compare")
	_, err := gradeService.ParseAndCompare()
	if err != nil {
		slog.Error("Initial parse and compare failed", "error", err)
	}

	bot := telegram.NewTelegramBot(cfg.TelegramConfig, gradeService)
	slog.Info("Bot started")

	go bot.RunBackSync()
	go bot.RunHandler()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Received shutdown signal, exiting...")
	time.Sleep(1 * time.Second)
}
