package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/config"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/scheduler"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/service"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/storage"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/telegram"
	"github.com/TheTeemka/telegram_bot_moodle_grades/pkg/logging"
)

func main() {
	debugFlag := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()
	logging.SetSlog(*debugFlag)

	cfg := config.Load()

	slog.Info("Starting telegram bot", "debug", *debugFlag)

	fetcher := service.NewMoodleFetcher(cfg.MoodleConfig)
	csvWriter := storage.NewCSVWriter(cfg.CsvFilesDir)
	gradeService := service.NewGradeService(fetcher, csvWriter, badTitles)

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

var badTitles = []string{
	"Academic Integrity: The Key to Your Success",
	"Workplace Instruction - SEDS / Жұмыс орнындағы нұсқама - ИСҒМ / Инструктаж на рабочем месте - ШИЦН",
	"University Security / Crisis Training",
	"Programming for Scientists and Engineers-Lecture,Section-5-Fall 2024",
	"Programming for Scientists and Engineers-CompLab,Section-2-Fall 2024",
	"Discrete Mathematics-Lecture,Section-4-Spring 2025",
	"Physics II for Scientists and Engineers with Laboratory-Recitation-S-1-12-Merged",
	"Calculus II-Lecture,Section-2-Spring 2025",
	"Physics II for Scientists and Engineers with Laboratory-Lecture,Section-1-Spring 2025",
	"History of Kazakhstan-Seminar,Section-1-Spring 2025",
	"History of Kazakhstan-Lecture,Section-1-Spring 2025",
	"Calculus II-Recitation,Section-1-Spring 2025",
	"Physics II for Scientists and Engineers with Laboratory-PhysLab,Section-5-Spring 2025",
	"Performance and Data Structures-Lab,Section-4-Spring 2025",
	"Performance and Data Structures-Lecture,Section-1-Spring 2025",
	"Linear Algebra with Applications-Lecture,Section-1-Summer 2025",
	"Linear Algebra with Applications-Recitation,Section-1-Summer 2025",
	"Sandbox course",
}
