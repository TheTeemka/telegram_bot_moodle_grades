package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/config"
)

func main() {
	cfg := config.Load()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	fetcher := NewMoodleFetcher(
		cfg.MoodleLoginSite,
		cfg.MoodleGradeSite,
		cfg.MoodleUser,
		cfg.MoodlePass,
	)

	parser := NewParser(fetcher)

	bot := NewTelegramBot(cfg.TelegramToken, cfg.TelegramID, parser.GetLastTimeParsed)

	syncChan := make(chan struct{})
	go bot.RunInputWorker(syncChan)

	output := make(chan string)
	go bot.RunOutputWorker(output)

	parser.parseAndCompare(output)

	done := make(chan bool)

	const dur = 3 * time.Hour
	go func() {
		ticker := time.NewTicker(dur)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				parser.parseAndCompare(output)
				slog.Info("Automatic sync triggered")
			case <-syncChan:
				parser.parseAndCompare(output)
				ticker.Reset(dur)
				slog.Info("Manual sync triggered, reset ticker")
			case <-done:
				slog.Info("Stopping background parsing")
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	slog.Info("Received shutdown signal, exiting...")
	close(done)
	time.Sleep(1 * time.Second)
}
