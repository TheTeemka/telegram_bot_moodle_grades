package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/config"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/scheduler"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/service"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/storage"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/telegram"
)

func main() {
	debugFlag := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()
	setSlog(*debugFlag)
	cfg := config.Load()

	slog.Info("Starting telegram bot", "debug", *debugFlag)

	fetcher := service.NewMoodleFetcher(cfg.MoodleConfig)
	csvWriter := storage.NewCSVWriter(cfg.CsvFilesDir)
	gradeService := service.NewGradeService(fetcher, csvWriter)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	bot := telegram.NewTelegramBot(cfg.TelegramConfig, gradeService)
	wg.Go(func() {
		bot.Run(ctx)
	})
	slog.Info("Bot started")

	scheduler := scheduler.NewSyncScheduler(cfg.SyncInterval, bot.HandleSync)
	wg.Go(func() {
		scheduler.Run(ctx)
	})
	slog.Info("Background sync started", "interval", cfg.SyncInterval.String())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Received shutdown signal, exiting...")
	cancel()
	wg.Wait()
	slog.Info("Shutdown down.")
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
