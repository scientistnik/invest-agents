package exchanges

import (
	"github.com/scientistnik/invest-agents/internal/app"
	"github.com/scientistnik/invest-agents/internal/app/domain"
)

type ExchangeId int

const (
	_                        = iota
	CurrencyId    ExchangeId = iota
	MaxExchangeId ExchangeId = iota
)

type AppExchange struct{}

var _ app.AppExchange = (*AppExchange)(nil)

func (ae AppExchange) GetExchangeByJson(exchangeId int, data []byte) domain.Exchange {
	switch exchangeId {
	case int(CurrencyId):
		return GetCurrencyByJson(data)
	}
	return nil
}
