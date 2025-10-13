package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/config"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/service"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/telegram"
)

func main() {
	debugFlag := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	setSlog(*debugFlag)

	cfg := config.Load()

	slog.Info("Starting telegram bot")
	fetcher := service.NewMoodleFetcher(cfg.MoodleConfig)
	gradeService := service.NewGradeService(fetcher)

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

func setSlog(debug bool) {
	l := slog.LevelInfo
	if debug {
		l = slog.LevelDebug
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     l,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.SourceKey {
				return a
			}

			switch v := a.Value.Any().(type) {
			case *slog.Source:
				if v != nil {
					short := filepath.Base(v.File)
					a.Value = slog.StringValue(fmt.Sprintf("%s:%d", short, v.Line))
				}
			case slog.Source:
				short := filepath.Base(v.File)
				a.Value = slog.StringValue(fmt.Sprintf("%s:%d", short, v.Line))
			default:
				// Fallback: shorten the string representation
				s := a.Value.String()
				if s != "" {
					a.Value = slog.StringValue(filepath.Base(s))
				}
			}
			return a
		},
	})

	slog.SetDefault(slog.New(h))
}
