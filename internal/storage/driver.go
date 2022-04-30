package storage

import (
	"github.com/scientistnik/invest-agents/internal/app"
	"github.com/scientistnik/invest-agents/internal/app/domain"
)

type Driver interface {
	migrateRollUp() error
	migrateRollDown() error
	connect() error
	disconnect() error
	userGetOrCreate(links app.UserLinks) (*domain.User, error)
	agentFind(filter app.AgentFilter) ([]domain.Agent, error)
	agentCreate(agent domain.Agent) (*domain.Agent, error)
	agentSetStatus(agent *domain.Agent, status domain.AgentStatus) error
	agentUpdateData(agent *domain.Agent, data []byte) error
	getAgentExchanges(agentId int64) ([]app.ExchangeData, error)
}
