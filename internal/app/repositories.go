package app

import (
	"github.com/scientistnik/invest-agents/internal/app/domain"
)

type AgentRepo struct {
	storage *AppStorage
}

var _ domain.AgentRepo = (*AgentRepo)(nil)

func (a AgentRepo) FindAgents(active bool) ([]domain.Agent, error) {
	return (*a.storage).FindAgents(AgentFilter{Status: domain.ActiveAgentStatus})
}

// type StrategyRepo struct {
// 	storage *AppStorage
// }

// func (sr StrategyRepo) GetStrategyData(agentId string) []byte {
// 	return (*sr.storage).GetStrategyData(agentId)
// }

type StorageRepo struct {
	storage *AppStorage
}

var _ domain.StorageRepo = (*StorageRepo)(nil)

func (s StorageRepo) GetAgentStorage(agent domain.Agent) interface{} {
	return (*s.storage).GetAgentStorage(agent) // !!!
}

type ExchangeRepo struct {
	storage  *AppStorage
	exchange *AppExchange
}

var _ domain.ExchangeRepo = (*ExchangeRepo)(nil)

func (e ExchangeRepo) GetAgentExchanges(agentId int64) ([]domain.Exchange, error) {
	exchs, err := (*e.storage).GetAgentExchanges(agentId)
	if err != nil {
		return nil, err
	}

	exchanges := []domain.Exchange{}
	for _, exch := range exchs {
		exchanges = append(exchanges, (*e.exchange).GetExchangeByJson(exch.Id, exch.Data))
	}

	return exchanges, nil
}
