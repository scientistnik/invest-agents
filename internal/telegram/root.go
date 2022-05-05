package telegram

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/scientistnik/invest-agents/internal/app"
)

func Start(ctx context.Context, token string, actions *app.Actions) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("telegram error, %w", err)
	}

	bot.Debug = true

	//log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				return nil
			}

			if update.Message == nil { // ignore any non-Message updates
				continue
			}

			// if !update.Message.IsCommand() { // ignore any non-command Messages
			// 	continue
			// }

			//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			if update.Message.IsCommand() {
				// Create a new MessageConfig. We don't have text yet,
				// so we leave it empty.

				// Extract the command from the Message.
				switch update.Message.Command() {
				case "start":
					msg.Text = "Hello"
				case "help":
					msg.Text = "I understand /sayhi and /status."
				case "sayhi":
					msg.Text = "Hi :)"
				case "status":
					msg.Text = "I'm ok."
				default:
					msg.Text = "I don't know that command"
				}

			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID
				//bot.Send(msg)
			}

			if _, err := bot.Send(msg); err != nil {
				fmt.Printf("telegram error: %#v\n", err)
			}

		case <-ctx.Done():
			return nil
		}
	}
}
