package domain

type Repos struct {
	Agent AgentRepo
	//Strategy StrategyRepo
	Storage  StorageRepo
	Exchange ExchangeRepo
	Logger   LoggerRepo
}

func StartAgents(repos Repos) error {
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

		err = strategy.Run(storage, exchanges, logger)
		if err != nil {
			return err
		}
	}

	return nil
}
