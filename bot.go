package main

import (
	"fmt"
	"log/slog"
	"time"

	tapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	bot             *tapi.BotAPI
	targetID        int64
	getLastTimeSync func() time.Time
}

func NewTelegramBot(token string, targetID int64, getLastTimeSync func() time.Time) *TelegramBot {
	botAPI, err := tapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	bot := &TelegramBot{
		bot:             botAPI,
		targetID:        targetID,
		getLastTimeSync: getLastTimeSync,
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

func (b *TelegramBot) SendToTarget(msg string) error {
	message := tapi.NewMessage(b.targetID, msg)
	message.ParseMode = tapi.ModeHTML
	_, err := b.bot.Send(message)
	return err
}

func (b *TelegramBot) RunInputWorker(syncChan chan<- struct{}) {
	u := tapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			switch update.Message.Command() {
			case "start":
				err := b.SendToTarget("Bot is running!")
				if err != nil {
					slog.Error("Failed to send message", "error", err)
				}
			case "sync":
				select {
				case syncChan <- struct{}{}:
					err := b.SendToTarget("Sync triggered!")
					if err != nil {
						slog.Error("Failed to send sync message", "error", err)
					}
				default:
					// Channel full, ignore to avoid blocking
					err := b.SendToTarget("Sync already in progress.")
					if err != nil {
						slog.Error("Failed to send sync message", "error", err)
					}
				}
			case "status":
				msg := fmt.Sprintf("Last parsed at: %s", b.getLastTimeSync().Format("2006-01-02 15:04:05"))
				err := b.SendToTarget(msg)
				if err != nil {
					slog.Error("Failed to send status", "error", err)
				}
			}
		}
	}

}
