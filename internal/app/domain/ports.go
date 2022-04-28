package domain

import "github.com/shopspring/decimal"

type Exchange interface {
	Balances(assets []string) ([]Balance, error)
	GetOpenOrders() ([]Order, error)
	GetHistoryOrders(pairs []Pair) ([]Order, error)
	LastPrice(pair Pair) (decimal.Decimal, error)
	Buy(pair Pair, amount decimal.Decimal) (*Order, error)
	Sell(pair Pair, amount decimal.Decimal, price decimal.Decimal) (*Order, error)
	GetOrderFee(pair Pair, amount decimal.Decimal) (*decimal.Decimal, error)
}

type Storage interface{}

type Logger interface {
	Info(message string)
	Warn(message string)
	Error(message string)
	Debug(message string)
}

type UserFindFilter struct {
	Ids   []string
	Links map[string]interface{}
}

type UserRepo interface {
	Find(filter UserFindFilter) ([]User, error)
	Create(linkName string, linkValue interface{}) (*User, error)
	Update(userId string, newName string) (*User, error)
}

type AgentRepo interface {
	//GetActiveAgents() []Agent
	FindAgents(active bool) ([]Agent, error)
}

type StrategyRepo interface {
	GetStrategyData(agentId string) []byte
}

type StorageRepo interface {
	GetAgentStorage(agent Agent) interface{}
}

type ExchangeRepo interface {
	GetAgentExchanges(agentId int64) ([]Exchange, error)
}

type LoggerRepo interface {
	New(agentId int64) Logger
}
