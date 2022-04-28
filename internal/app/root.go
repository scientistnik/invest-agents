package app

import (
	"github.com/scientistnik/invest-agents/internal/app/domain"
)

type ExchangeData struct {
	Id   int
	Data []byte
}

type Actions struct {
	Storage AppStorage
}

type UserLinks struct {
	Telegram int64 `json:"telegram"`
}

func (a Actions) UserGetOrCreate(name string, links UserLinks) (*domain.User, error) {
	return a.Storage.UserGetOrCreate(links)
}

func (a Actions) GetUserAgents(user domain.User) ([]domain.Agent, error) {
	return a.Storage.FindAgents(AgentFilter{UserId: user.Id})
}

func (a Actions) FindAgents(filter AgentFilter) ([]domain.Agent, error) {
	return a.Storage.FindAgents(filter)
}

func (a Actions) AgentCreate(user domain.User, strategyId domain.StrategyId) (*domain.Agent, error) {
	return a.Storage.AgentSave(domain.Agent{UserId: user.Id, Status: domain.DisableAgentStatus, StrategyId: strategyId})
}

func (a Actions) StartAgents(exchanges AppExchange, appLogger domain.LoggerRepo) error {
	return domain.StartAgents(domain.Repos{
		Agent: AgentRepo{storage: &a.Storage},
		//Strategy: StrategyRepo{storage: &storage},
		Storage:  StorageRepo{storage: &a.Storage},
		Exchange: ExchangeRepo{storage: &a.Storage, exchange: &exchanges},
		Logger:   appLogger,
	})
}
