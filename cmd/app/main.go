package main

import (
	"fmt"
	"github.com/scientistnik/invest-agents/internal/app"
	"github.com/scientistnik/invest-agents/internal/app/domain"
	"github.com/scientistnik/invest-agents/internal/exchanges"
	"github.com/scientistnik/invest-agents/internal/loggers"
	"github.com/scientistnik/invest-agents/internal/storage"
)

func main() {
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

	actions := app.Actions{Storage: appStorage}
	user, err := actions.UserGetOrCreate("test", app.UserLinks{Telegram: 12})
	if err != nil {
		fmt.Printf("error in userGetOrCreate: %#v\n", err)
		return
	}
	fmt.Printf("User: %#v\n", user)

	agents, err := actions.FindAgents(app.AgentFilter{UserId: user.Id})
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	var agent *domain.Agent
	if len(agents) == 0 {
		agent, err = actions.AgentCreate(*user, domain.SimpleStratedy)
		if err != nil {
			fmt.Printf("%#v\n", err)
		}
	} else {
		agent = &agents[0]
	}

	fmt.Printf("Agent: %#v\n", agent)

	err = actions.StartAgents(exchanges.AppExchange{}, loggers.ConstructorConsoleLogger{})
	if err != nil {
		fmt.Printf("\n%#v\n", err)
	}
}
