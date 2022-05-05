package app

import (
	"context"
	"github.com/scientistnik/invest-agents/internal/app/domain"
)

type ExchangeData struct {
	Id   int
	Data []byte
}

type Actions struct {
	storage  AppStorage
	exchange AppExchange
	logger   domain.LoggerRepo
}

func GetAppActions(storage AppStorage, exchange AppExchange, appLogger domain.LoggerRepo) *Actions {
	return &Actions{storage: storage, exchange: exchange, logger: appLogger}
}

type UserLinks struct {
	Telegram int64 `json:"telegram"`
}

func (a Actions) UserGetOrCreate(name string, links UserLinks) (*domain.User, error) {
	return a.storage.UserGetOrCreate(links)
}

type ExchangeFilter struct {
	UserId         int64
	ExchangeNumber int
}

func (a Actions) FindExchanges(filter ExchangeFilter) ([]ExchangeData, error) {
	return a.storage.FindExchanges(filter)
}

func (a Actions) AddExchange(user domain.User, exchangeNumber int, data []byte) error {
	return a.storage.AddExchange(user.Id, exchangeNumber, data)
}

func (a Actions) GetUserAgents(user domain.User) ([]domain.Agent, error) {
	return a.storage.FindAgents(AgentFilter{UserId: user.Id})
}

func (a Actions) FindAgents(filter AgentFilter) ([]domain.Agent, error) {
	return a.storage.FindAgents(filter)
}

func (a Actions) AgentCreate(user domain.User, strategyId domain.StrategyId, data []byte, exchanges []ExchangeData) (*domain.Agent, error) {
	agent, err := a.storage.AgentSave(domain.Agent{UserId: user.Id, Status: domain.DisableAgentStatus, StrategyId: strategyId, StrategyData: data})
	if err != nil {
		return nil, err
	}

	err = a.storage.AgentAddExchange(agent, exchanges)
	if err != nil {
		return nil, err
	}

	return agent, err
}

func (a Actions) AgentSetStatus(agent *domain.Agent, status domain.AgentStatus) error {
	return a.storage.AgentSetStatus(agent, status)
}

func (a Actions) AgentUpdateData(agent *domain.Agent, data []byte) error {
	return a.storage.AgentUpdateData(agent, data)
}

func (a Actions) StartAgents(ctx context.Context) error {
	return domain.StartAgents(ctx, domain.Repos{
		Agent:    AgentRepo{storage: &a.storage},
		Storage:  StorageRepo{storage: &a.storage},
		Exchange: ExchangeRepo{storage: &a.storage, exchange: &a.exchange},
		Logger:   a.logger,
	})
}
