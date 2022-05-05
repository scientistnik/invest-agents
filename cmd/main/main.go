package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/scientistnik/invest-agents/internal/app"
	"github.com/scientistnik/invest-agents/internal/exchanges"
	"github.com/scientistnik/invest-agents/internal/loggers"
	"github.com/scientistnik/invest-agents/internal/storage"
	"github.com/scientistnik/invest-agents/internal/telegram"
	"os"
	"os/signal"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")

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

	wg.Add(1)
	go func() {
		defer wg.Done()

		err = actions.StartAgents(ctx)
		if err != nil {
			fmt.Printf("\n%#v\n", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		telegram.Start(ctx, telegramToken, actions)
	}()

	cancelCannel := make(chan os.Signal, 1)
	signal.Notify(cancelCannel, os.Interrupt)

	<-cancelCannel

	cancel()

	wg.Wait()
}
