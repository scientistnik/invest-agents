package app

import (
	"context"
	"fmt"
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
	repos    domain.Repos
}

func GetAppActions(storage AppStorage, exchange AppExchange, appLogger domain.LoggerRepo) *Actions {
	return &Actions{
		storage:  storage,
		exchange: exchange,
		logger:   appLogger,
		repos: domain.Repos{
			Agent:    AgentRepo{storage: &storage},
			Storage:  StorageRepo{storage: &storage},
			Exchange: ExchangeRepo{storage: &storage, exchange: &exchange},
			Logger:   appLogger,
		},
	}
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
	return domain.StartAgents(ctx, a.repos)
}

type AgentInfo struct {
	Name         string
	StrategyName string
	Exchanges    []string
	Parameters   []domain.StrategyParameter
}

func (a Actions) GetAgentInfo(agent domain.Agent) *AgentInfo {
	strategy := domain.GetStrategyFromJson(agent.StrategyId, agent.StrategyData)

	exchanges, err := a.repos.Exchange.GetAgentExchanges(agent.Id)
	if err != nil {
		return nil
	}

	exchangeNames := []string{}
	for _, exchange := range exchanges {
		exchangeNames = append(exchangeNames, exchange.Name())
	}

	return &AgentInfo{
		Name:         fmt.Sprintf("Agent %d", agent.Id),
		StrategyName: strategy.Name(),
		Exchanges:    exchangeNames,
		Parameters:   strategy.Parameters(),
	}
}
