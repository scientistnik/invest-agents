package app

import (
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

func (a Actions) FindUserExchanges(user domain.User, exchangeId int) []ExchangeData {
	return nil
}

func (a Actions) UserAddExchange(user domain.User, exchangeId int, data []byte) error {
	return nil
}

func (a Actions) GetUserAgents(user domain.User) ([]domain.Agent, error) {
	return a.storage.FindAgents(AgentFilter{UserId: user.Id})
}

func (a Actions) FindAgents(filter AgentFilter) ([]domain.Agent, error) {
	return a.storage.FindAgents(filter)
}

func (a Actions) AgentCreate(user domain.User, strategyId domain.StrategyId, data []byte) (*domain.Agent, error) {
	return a.storage.AgentSave(domain.Agent{UserId: user.Id, Status: domain.DisableAgentStatus, StrategyId: strategyId, StrategyData: data})
}

func (a Actions) AgentSetStatus(agent *domain.Agent, status domain.AgentStatus) error {
	return a.storage.AgentSetStatus(agent, status)
}

func (a Actions) AgentUpdateData(agent *domain.Agent, data []byte) error {
	return a.storage.AgentUpdateData(agent, data)
}

func (a Actions) StartAgents() error {
	return domain.StartAgents(domain.Repos{
		Agent:    AgentRepo{storage: &a.storage},
		Storage:  StorageRepo{storage: &a.storage},
		Exchange: ExchangeRepo{storage: &a.storage, exchange: &a.exchange},
		Logger:   a.logger,
	})
}
