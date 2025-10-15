package telegram

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

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

func (b *TelegramBot) Send(chatID int64, msg string) error {
	message := tapi.NewMessage(chatID, msg)
	message.ParseMode = tapi.ModeHTML
	_, err := b.bot.Send(message)
	return err
}

func (b *TelegramBot) SendToTarget(msg string) error {
	return b.Send(b.targetID, msg)
}

func (b *TelegramBot) RunHandler() {
	u := tapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)

	const NumWorker = 3
	for range NumWorker {
		go b.runHandlerWorker(updates)
	}
}

func (b *TelegramBot) runHandlerWorker(updates <-chan tapi.Update) {
	for update := range updates {
		if !b.IsFromMe(update) {
			continue
		}

		if update.Message != nil {
			switch update.Message.Command() {
			case "start":
				b.HandleStart()
			case "sync":
				err := b.SendToTarget("Manual sync triggered")
				if err != nil {
					slog.Error("Failed to send sync message", "error", err)
				}

				err = b.HandleSync()
				if err != nil {
					continue
				}

				err = b.SendToTarget("Manual sync finished")
				if err != nil {
					slog.Error("Failed to send sync message", "error", err)
				}
			case "status":
				b.HandlerStatus()
			}
		}
	}
}
func (b *TelegramBot) RunBackSync() {
	const dur = 3 * time.Hour
	go func() {
		ticker := time.NewTicker(dur)
		defer ticker.Stop()
		for range ticker.C {
			b.HandleSync()
			slog.Debug("Automatic sync triggered")
		}

	}()
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

func (b *TelegramBot) HandlerStatus() {
	msg := fmt.Sprintf("Last parsed at: %s", b.gradeService.LastTimeParsed.Format("2006-01-02 15:04:05"))
	err := b.SendToTarget(msg)
	if err != nil {
		slog.Error("Failed to send status", "error", err)
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
func (b *TelegramBot) StartMessage() {
	err := b.SendToTarget("✳️ Bot has started!")
	if err != nil {
		slog.Error("Failed to send start message", "error", err)
	}
}

func (b *TelegramBot) DeadMessage() {
	err := b.SendToTarget("☠️ Bot shutting down")
	if err != nil {
		slog.Error("Failed to send shutdown message", "error", err)
	}
}

func (b *TelegramBot) Spam(d time.Duration) {
	ch := make(chan struct{}, 100)
	go func() {
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		for range ticker.C {
			ch <- struct{}{}
		}
	}()
	for range ch {
		err := b.SendToTarget("Spam message" + time.Now().Format(time.RFC3339))
		if err != nil {
			slog.Error("Failed to send spam message", "error", err)
		}
	}

}
