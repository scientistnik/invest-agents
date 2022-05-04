package exchanges

import (
	"encoding/json"
	"github.com/scientistnik/invest-agents/internal/app/domain"

	currencycom "github.com/scientistnik/currency.com"
	"github.com/shopspring/decimal"
	"strings"
)

type Currency struct {
	api currencycom.RestAPI
}

var _ domain.Exchange = (*Currency)(nil)

type CurrencyData struct {
	ApiKey string `json:"api_key"`
	Secret string `json:"secret"`
}

func GetCurrencyFromJson(data []byte) (*Currency, error) {
	var c CurrencyData

	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	return &Currency{api: *currencycom.NewRestAPI(c.ApiKey, c.Secret, currencycom.DEFAULT_ENDPOINT)}, nil
}

func GetCurrencyToJson(cd CurrencyData) ([]byte, error) {
	return json.Marshal(&cd)
}

func convertOrderStatusStringToInt(orderStatus string) domain.OrderStatus {
	//CANCELED, EXPIRED, FILLED, NEW, PARTIALLY_FILLED, PENDING_CANCEL, REJECTED
	switch orderStatus {
	case "FILLED":
		return domain.FillOrderStatus
	case "CANCELED":
		return domain.CanceledOrderStatus
	default:
		return domain.PendingOrderStatus
	}
}

func convertPairStringToStruct(symbol string) domain.Pair {
	assets := strings.Split(symbol, "/")
	return domain.Pair{BaseAsset: assets[0], QuoteAsset: assets[1]}
}

func convertPairStructToString(pair domain.Pair) string {
	return pair.BaseAsset + "/" + pair.QuoteAsset
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

func (c *Currency) GetOpenOrders(filter *domain.OrderFilter) ([]domain.Order, error) {
	var orderFilter currencycom.PositionHistoryRequest

	if filter != nil && len(filter.Pairs) == 1 {
		pair := filter.Pairs[0]
		orderFilter = currencycom.PositionHistoryRequest{Symbol: pair.BaseAsset + "/" + pair.QuoteAsset}
	}

	currencyOrders, err := c.api.ListOfOpenOrder(&orderFilter)
	if err != nil {
		return nil, err
	}

	var orders []domain.Order
	for _, currOrder := range currencyOrders {
		orderStatus := convertOrderStatusStringToInt(currOrder.Status)
		orderPair := convertPairStringToStruct(currOrder.Symbol)

		if filter != nil {
			excludeOrder := true

			if len(filter.Ids) > 0 {
				for _, id := range filter.Ids {
					if currOrder.OrderId == id {
						excludeOrder = false
						break
					}
				}

				if excludeOrder {
					continue
				}
			}

			if len(filter.Statuses) > 0 {
				excludeOrder = true

				for _, status := range filter.Statuses {
					if orderStatus == status {
						excludeOrder = false
						break
					}
				}

				if excludeOrder {
					continue
				}
			}

			if len(filter.Pairs) > 1 {
				excludeOrder = true

				for _, pair := range filter.Pairs {
					if orderPair.BaseAsset == pair.BaseAsset && orderPair.QuoteAsset == pair.QuoteAsset {
						excludeOrder = false
						break
					}
				}

				if excludeOrder {
					continue
				}
			}

		}

		price, err := decimal.NewFromString(currOrder.Price)
		if err != nil {
			return nil, err
		}

		amount, err := decimal.NewFromString(currOrder.OrigQty)
		if err != nil {
			return nil, err
		}

		orders = append(orders, domain.Order{
			Id:     currOrder.OrderId,
			Status: orderStatus,
			Price:  price,
			Amount: amount,
			Pair:   orderPair,
			//Commission: 0,
		})
	}

	return orders, nil
}

func (c *Currency) GetHistoryOrders(pairs []domain.Pair) ([]domain.Order, error) {
	var filter currencycom.AllMyTradesRequest
	if len(pairs) == 1 {
		filter = currencycom.AllMyTradesRequest{Symbol: convertPairStructToString(pairs[0])}
	}

	trades, err := c.api.ListOfTrades(&filter)
	if err != nil {
		return nil, err
	}

	var orders []domain.Order
	for _, trade := range trades {

		if len(pairs) > 1 {
			excludeTrade := true

			for _, pair := range pairs {
				if convertPairStructToString(pair) == trade.Symbol {
					excludeTrade = false
					break
				}
			}

			if excludeTrade {
				continue
			}
		}

		pair := convertPairStringToStruct(trade.Symbol)

		price, err := decimal.NewFromString(trade.Price)
		if err != nil {
			return nil, err
		}

		amount, err := decimal.NewFromString(trade.Qty)
		if err != nil {
			return nil, err
		}

		comission, err := decimal.NewFromString(trade.Commission)
		if err != nil {
			return nil, err
		}

		orders = append(orders, domain.Order{
			Id:         trade.Id,
			Status:     domain.FillOrderStatus,
			Price:      price,
			Amount:     amount,
			Pair:       pair,
			Commission: domain.Balance{Asset: trade.CommissionAsset, Amount: comission},
		})
	}

	return orders, nil
}

func (c *Currency) LastPrice(pair domain.Pair) (decimal.Decimal, error) {
	ticker, err := currencycom.PriceChange(&currencycom.BySymbolRequest{Symbol: convertPairStructToString(pair)})
	if err != nil {
		return decimal.New(0, 0), err
	}

	return decimal.NewFromString(ticker.LastPrice)
}

func (c *Currency) Buy(pair domain.Pair, amount decimal.Decimal) (*domain.Order, error) {
	quantity, _ := amount.Float64()

	result, err := c.api.CreateOrder(&currencycom.CreateOrderRequest{
		Symbol:   convertPairStructToString(pair),
		Quantity: quantity,
		Type:     "MARKET",
		Side:     "BUY",
	})
	if err != nil {
		return nil, err
	}

	price, err := decimal.NewFromString(result.Price)
	if err != nil {
		return nil, err
	}

	executedQty, err := decimal.NewFromString(result.ExecutedQty)
	if err != nil {
		return nil, err
	}

	commission, err := c.GetOrderFee(pair, amount, price)
	if err != nil {
		return nil, err
	}

	order := domain.Order{
		Id:         result.OrderId,
		Status:     domain.PendingOrderStatus,
		Price:      price,
		Amount:     executedQty,
		Pair:       convertPairStringToStruct(result.Symbol),
		Commission: commission,
	}

	return &order, nil
}

func (c *Currency) Sell(pair domain.Pair, amount decimal.Decimal, price decimal.Decimal) (*domain.Order, error) {
	quantity, _ := amount.Float64()
	floatPrice, _ := price.Float64()

	result, err := c.api.CreateOrder(&currencycom.CreateOrderRequest{
		Symbol:   convertPairStructToString(pair),
		Quantity: quantity,
		Type:     "LIMIT",
		Side:     "SELL",
		Price:    floatPrice,
	})
	if err != nil {
		return nil, err
	}

	resPrice, err := decimal.NewFromString(result.Price)
	if err != nil {
		return nil, err
	}

	executedQty, err := decimal.NewFromString(result.ExecutedQty)
	if err != nil {
		return nil, err
	}

	commission, err := c.GetOrderFee(pair, amount, price)
	if err != nil {
		return nil, err
	}

	order := domain.Order{
		Id:         result.OrderId,
		Status:     domain.PendingOrderStatus,
		Price:      resPrice,
		Amount:     executedQty,
		Pair:       convertPairStringToStruct(result.Symbol),
		Commission: commission,
	}

	return &order, nil
}

func (c *Currency) CancelOrder(orderId string, pair domain.Pair) error {
	_, err := c.api.CancelOrder(&currencycom.CancelOrderRequest{OrderId: orderId, Symbol: convertPairStructToString(pair)})
	return err
}

func (c *Currency) GetOrderFee(pair domain.Pair, amount decimal.Decimal, price decimal.Decimal) (domain.Balance, error) {
	fee, err := decimal.NewFromString("0.002")
	return domain.Balance{Asset: pair.QuoteAsset, Amount: amount.Mul(price).Mul(fee).RoundUp(2)}, err
}
