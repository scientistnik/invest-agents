package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/scientistnik/invest-agents/internal/app"
	"github.com/scientistnik/invest-agents/internal/app/domain"
	"github.com/scientistnik/invest-agents/internal/exchanges"
	"github.com/scientistnik/invest-agents/internal/loggers"
	"github.com/scientistnik/invest-agents/internal/storage"
	"github.com/shopspring/decimal"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	appStorage, err := storage.GetSqliteAppStorage("database.db")
	if err != nil {
		fmt.Println("error in creation", err)
		return
	}

	err = appStorage.Connect()
	if err != nil {
		fmt.Println("error in connection", err)
		return
	}

	defer appStorage.Disconnect()

	actions := app.GetAppActions(appStorage, exchanges.AppExchange{}, loggers.ConstructorConsoleLogger{Color: true})
	user, err := actions.UserGetOrCreate("test", app.UserLinks{Telegram: 12})
	if err != nil {
		fmt.Printf("error in userGetOrCreate: %#v\n", err)
		return
	}
	fmt.Printf("User: %#v\n", user)

	exchs, err := actions.FindExchanges(app.ExchangeFilter{UserId: user.Id, ExchangeNumber: int(exchanges.CurrencyId)})
	if err != nil {
		fmt.Printf("error in FindExchanges: %#v\n", err)
		return
	}
	if len(exchs) == 0 {
		data, err := exchanges.GetCurrencyToJson(exchanges.CurrencyData{ApiKey: apiKey, Secret: secretKey})
		if err != nil {
			fmt.Printf("%#v\n", err)
			return
		}

		err = actions.AddExchange(*user, int(exchanges.CurrencyId), data)
		if err != nil {
			fmt.Printf("%#v\n", err)
			return
		}

		exchs, _ = actions.FindExchanges(app.ExchangeFilter{UserId: user.Id, ExchangeNumber: int(exchanges.CurrencyId)})
	}

	agents, err := actions.FindAgents(app.AgentFilter{UserId: user.Id})
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	var agent *domain.Agent
	if len(agents) == 0 {
		simpleStrategyJson, err := domain.SimpleStratedyToJson(&domain.SimpleStrategy{
			Pair:            domain.Pair{BaseAsset: "BTC", QuoteAsset: "USD"},
			BaseQuality:     decimal.NewFromFloat(0.001),
			MaxTrades:       10,
			ProfitPercent:   decimal.NewFromFloat(0.005),
			FarPricePercent: decimal.NewFromFloat(0.01),
		})
		if err != nil {
			fmt.Printf("%#v\n", err)
		}

		agent, err = actions.AgentCreate(*user, domain.SimpleStratedy, simpleStrategyJson, exchs)
		if err != nil {
			fmt.Printf("%#v\n", err)
		}
	} else {
		agent = &agents[0]
	}

	fmt.Printf("Agent: %#v\n", agent)

	actions.AgentSetStatus(agent, domain.ActiveAgentStatus)
	// fmt.Printf("Update agent: %#v\n", agent)

	// fmt.Printf("%#v\n", string(simpleStrategyJson))
	// err = actions.AgentUpdateData(agent, simpleStrategyJson)
	// if err != nil {
	// 	fmt.Printf("Error in %#v\n", err)
	// 	return
	// }

	err = actions.StartAgents()
	if err != nil {
		fmt.Printf("\n%#v\n", err)
	}
}
