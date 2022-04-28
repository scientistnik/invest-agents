package storage

import (
	"github.com/scientistnik/invest-agents/internal/app"
	"github.com/scientistnik/invest-agents/internal/app/domain"

	packr "github.com/gobuffalo/packr/v2"
	migrate "github.com/rubenv/sql-migrate"
)

var migrations *migrate.PackrMigrationSource = &migrate.PackrMigrationSource{
	Box: packr.New("migrations", "./migrations"),
}

type AppStorage struct {
	driver Driver
}

var _ app.AppStorage = (*AppStorage)(nil)

func (as AppStorage) Connect() error {
	return as.driver.connect()
}

func (as AppStorage) Disconnect() error {
	return as.driver.disconnect()
}

func (as AppStorage) UserGetOrCreate(links app.UserLinks) (*domain.User, error) {
	return as.driver.userGetOrCreate(links)
}

func (as AppStorage) FindAgents(filter app.AgentFilter) ([]domain.Agent, error) {
	agents, err := as.driver.agentFind(filter)
	if err != nil {
		return nil, err
	}

	return agents, nil
}

func (as AppStorage) AgentSave(agent domain.Agent) (*domain.Agent, error) {
	return as.driver.agentCreate(agent)
}

func (as AppStorage) GetAgentStorage(agent domain.Agent) interface{} {
	switch agent.StrategyId {
	case domain.SimpleStratedy:
		return SimpleStorage{agent: agent}
	}

	return nil
}

func (as AppStorage) GetAgentExchanges(agentId int64) ([]app.ExchangeData, error) {
	return as.driver.getAgentExchanges(agentId)
}