package app

import "github.com/scientistnik/invest-agents/internal/app/domain"

type AgentFilter struct {
	Id     int64
	Status domain.AgentStatus
	UserId int64
}

type AppStorage interface {
	// User
	UserGetOrCreate(links UserLinks) (*domain.User, error)
	// Agent
	FindAgents(filter AgentFilter) ([]domain.Agent, error)
	AgentSave(agent domain.Agent) (*domain.Agent, error)
	AgentSetStatus(agent *domain.Agent, status domain.AgentStatus) error
	AgentUpdateData(agent *domain.Agent, data []byte) error
	//GetStrategyData(agentId string) []byte
	GetAgentStorage(strategyId domain.Agent) interface{}
	GetAgentExchanges(agentId int64) ([]ExchangeData, error)
}

type AppExchange interface {
	GetExchangeByJson(exchangeId int, data []byte) domain.Exchange
}
