package domain

import (
	"context"
	"sync"
	"time"
)

type Repos struct {
	Agent AgentRepo
	//Strategy StrategyRepo
	Storage  StorageRepo
	Exchange ExchangeRepo
	Logger   LoggerRepo
}

func StartAgents(ctx context.Context, repos Repos) error {
	var wg sync.WaitGroup

	agents, err := repos.Agent.FindAgents(true)
	if err != nil {
		return err
	}

	for _, agent := range agents {
		strategy := GetStrategyFromJson(agent.StrategyId, agent.StrategyData)
		storage := repos.Storage.GetAgentStorage(agent)
		exchanges, err := repos.Exchange.GetAgentExchanges(agent.Id)
		if err != nil {
			return err
		}
		logger := repos.Logger.New(agent.Id)

		wg.Add(1)
		go func() {
			defer wg.Done()

			workCycle := true

			for workCycle {

				err = strategy.Run(ctx, storage, exchanges, logger)
				if err != nil {
					logger.Error(err.Error())
				}

				select {
				case <-ctx.Done():
					workCycle = false
				case <-time.After(60 * time.Second):
					continue
				}

			}
		}()
	}

	wg.Wait()
	return nil
}
