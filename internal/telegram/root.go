package telegram

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/scientistnik/invest-agents/internal/app"
	"github.com/scientistnik/invest-agents/internal/app/domain"
	"github.com/shopspring/decimal"
	"strconv"
	"strings"
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
			//msg.Text = "Hello"

			if update.Message.IsCommand() {
				// Create a new MessageConfig. We don't have text yet,
				// so we leave it empty.

				// Extract the command from the Message.
				switch update.Message.Command() {
				case "start":
					user, err := actions.UserGetOrCreate("", app.UserLinks{Telegram: update.Message.Chat.ID})
					if err != nil {
						msg.Text += fmt.Sprintf("%#v\n", err)
					}
					//msg.Text += fmt.Sprintf("%#v\n", user)
					agents, err := actions.GetUserAgents(*user)
					if err != nil {
						msg.Text += fmt.Sprintf("%#v\n", err)
					}
					//msg.Text += fmt.Sprintf("%#v\n", agents)
					msg.Text += fmt.Sprintf("You have %d active agents:\n", len(agents))
					for _, agent := range agents {
						agentInfo := actions.GetAgentInfo(agent)
						msg.Text += fmt.Sprintf(
							"Name: %s\nExchanges: %s\nStrategy:\n  Name: %s\n",
							agentInfo.Name,
							strings.Join(agentInfo.Exchanges, ","),
							agentInfo.StrategyName,
						)

						for _, param := range agentInfo.Parameters {
							var value string
							switch param.Type {
							case domain.BoolParameterType:
								val := param.Value.(bool)
								value = strconv.FormatBool(val)
							case domain.IntParameterType:
								val := param.Value.(int)
								value = strconv.FormatInt(int64(val), 10)
							case domain.StringParameterType:
								value = param.Value.(string)
							case domain.PercentParameterType:
								val := param.Value.(decimal.Decimal).Mul(decimal.NewFromInt(100))
								value = val.String() + " %"
							case domain.PairParameterType:
								val := param.Value.(domain.Pair)
								value = val.BaseAsset + "/" + val.QuoteAsset
							case domain.BalanceParameterType:
								val := param.Value.(domain.Balance)
								value = fmt.Sprintf("%s %s", val.Amount, val.Asset)
							}

							msg.Text += fmt.Sprintf("  %s: %s\n", param.Name, value)
						}
					}
				case "help":
					msg.Text = "I understand /start and /status."
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

			if len(msg.Text) > 0 {
				if _, err := bot.Send(msg); err != nil {
					fmt.Printf("telegram error: %#v\n", err)
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}
