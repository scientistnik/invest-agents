package domain

import "github.com/shopspring/decimal"

type AgentStatus int

const (
	ErrorAgentStatus   = iota
	ActiveAgentStatus  = iota
	DisableAgentStatus = iota
)

type Agent struct {
	Id           int64
	UserId       int64
	Status       AgentStatus
	StrategyId   StrategyId
	StrategyData []byte
	//Storage    interface{} //Storage
	//Exchange   []Exchange
	//Logger     Logger
}

type User struct {
	Id   int64
	Name string
}

type OrderStatus = int

const (
	_                           = iota
	OrderStatusFill OrderStatus = iota
)

type Order struct {
	Id         string
	Status     OrderStatus
	Price      decimal.Decimal
	Amount     decimal.Decimal
	Commission Balance
}
