package exchanges

import (
	"github.com/scientistnik/invest-agents/internal/app/domain"

	currencycom "github.com/scientistnik/currency.com"
	"github.com/shopspring/decimal"
)

type Currency struct {
	api currencycom.RestAPI
}

var _ domain.Exchange = (*Currency)(nil)

func NewCurrency(apiKey string, secret string) *Currency {
	return &Currency{api: *currencycom.NewRestAPI(apiKey, secret, currencycom.DEFAULT_ENDPOINT)}
}

func GetCurrencyByJson(data []byte) *Currency {
	return &Currency{}
}

func (c *Currency) Balances(assets []string) ([]domain.Balance, error) {
	res, err := c.api.AccountInfo(&currencycom.AccountRequest{ShowZeroBalance: true})
	if err != nil {
		return nil, err
	}

	balances := []domain.Balance{}
	for _, accBalance := range res.Balances {
		if len(assets) > 0 {
			for _, asset := range assets {
				if asset == accBalance.Asset {
					balances = append(balances, domain.Balance{Asset: asset, Amount: decimal.NewFromFloat(accBalance.Free)})
					break
				}
			}
		} else {
			balances = append(balances, domain.Balance{Asset: accBalance.Asset, Amount: decimal.NewFromFloat(accBalance.Free)})
		}
	}

	return balances, nil
}

func (c *Currency) GetOpenOrders() ([]domain.Order, error) {
	return nil, nil
}

func (c *Currency) GetHistoryOrders(pairs []domain.Pair) ([]domain.Order, error) {
	return nil, nil
}

func (c *Currency) LastPrice(pair domain.Pair) (decimal.Decimal, error) {
	return decimal.New(0, 0), nil
}

func (c *Currency) Buy(pair domain.Pair, amount decimal.Decimal) (*domain.Order, error) {
	return nil, nil
}

func (c *Currency) Sell(pair domain.Pair, amount decimal.Decimal, price decimal.Decimal) (*domain.Order, error) {
	return nil, nil
}

func (c *Currency) GetOrderFee(pair domain.Pair, amount decimal.Decimal) (*decimal.Decimal, error) {
	return nil, nil
}
