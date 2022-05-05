package domain

import "context"

type StrategyId int

const (
	_                            = iota
	SimpleStratedy    StrategyId = iota
	maxStrategyNumber StrategyId = iota
)

type Strategy interface {
	//NewFromJson(_json []byte) interface{}
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
