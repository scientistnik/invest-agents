package domain

type StrategyId int

const (
	_                            = iota
	SimpleStratedy    StrategyId = iota
	maxStrategyNumber StrategyId = iota
)

type Strategy interface {
	//NewFromJson(_json []byte) interface{}
	Run(storage interface{}, exchanges []Exchange, logger Logger) error
}

func GetStrategyFromJson(id StrategyId, data []byte) Strategy {
	switch id {
	case SimpleStratedy:
		strategy := NewSimpleStrategyFromJson(data)
		return strategy
	}

	return nil
}
