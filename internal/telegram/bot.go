package telegram

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/config"
	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/service"
	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	bot      *tapi.BotAPI
	targetID int64

	gradeService *service.GradeService
}

func NewTelegramBot(cfg config.TelegramConfig, gradeService *service.GradeService) *TelegramBot {
	botAPI, err := tapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		panic(err)
	}

	bot := &TelegramBot{
		bot:          botAPI,
		targetID:     cfg.TelegramID,
		gradeService: gradeService,
	}

	err = bot.SetCommands()
	if err != nil {
		panic(err)
	}

	return bot
}

func (b *TelegramBot) SetCommands() error {
	commandsConfig := tapi.NewSetMyCommands([]tapi.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "sync", Description: "Trigger a manual sync"},
		{Command: "status", Description: "Get the last sync time"},
	}...)

	_, err := b.bot.Request(commandsConfig)
	if err != nil {
		return err
	}
	return nil
}

func (b *TelegramBot) RunOutputWorker(output <-chan string) {
	for msg := range output {
		err := b.SendToTarget(msg)
		if err != nil {
			slog.Error("Failed to send message", "error", err)
		}
	}
}

func (b *TelegramBot) Run(ctx context.Context) {
	u := tapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)

	b.StartMessage()
	const NumWorker = 1
	for range NumWorker {
		go b.runHandlerWorker(ctx, updates)
	}

	<-ctx.Done()
	b.DeadMessage()
}

func (b *TelegramBot) runHandlerWorker(ctx context.Context, updates <-chan tapi.Update) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			if !b.IsFromMe(update) {
				continue
			}
			b.HandleUpdate(update)
		}
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
