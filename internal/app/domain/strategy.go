package domain

import "context"

type StrategyId int

const (
	_                            = iota
	SimpleStratedy    StrategyId = iota
	maxStrategyNumber StrategyId = iota
)

type Strategy interface {
	Name() string
	Parameters() []StrategyParameter
	ValidateParameter(StrategyParameter) bool
	Run(ctx context.Context, storage interface{}, exchanges []Exchange, logger Logger) error
}

func GetStrategyFromJson(id StrategyId, data []byte) Strategy {
	switch id {
	case SimpleStratedy:
		strategy, err := NewSimpleStrategyFromJson(data)
		if err != nil {
			panic(err)
		}
		return strategy
	}

	return nil
}

type StrategyParameterType = int

const (
	_                    = iota
	BoolParameterType    = iota
	IntParameterType     = iota
	StringParameterType  = iota
	PercentParameterType = iota
	PairParameterType    = iota
	BalanceParameterType = iota
)

type StrategyParameter struct {
	Type  StrategyParameterType
	Name  string
	Value interface{}
}
