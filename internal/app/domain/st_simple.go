package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

var Version string = "v0.3.0"

type SimpleStrategy struct {
	Pair            Pair            `json:"pair"`
	BaseQuality     decimal.Decimal `json:"base_quality"`
	MaxTrades       int             `json:"max_trades"`
	ProfitPercent   decimal.Decimal `json:"profit_percent"`
	FarPricePercent decimal.Decimal `json:"far_price_percent"`
}

var _ Strategy = (*SimpleStrategy)(nil)

type SimpleTradeStatus = int

const (
	_                                         = iota
	SimpleTradeStatusBuy    SimpleTradeStatus = iota
	SimpleTradeStatusSell   SimpleTradeStatus = iota
	SimpleTradeStatusFinish SimpleTradeStatus = iota
)

type SimpleTradeFilter struct {
	Statuses []SimpleTradeStatus
}

type SimpleStorage interface {
	//Storage
	GetTrades(filter *SimpleTradeFilter) ([]SimpleTrade, error)
	SaveTrade(trade *SimpleTrade) error
}

type SimpleTradeOrder struct {
	Datetime string
	OrderId  string
	Price    decimal.Decimal
	//amount     decimal.Decimal
	Commission Balance
}

type SimpleTrade struct {
	Id     int
	Status SimpleTradeStatus
	Amount decimal.Decimal
	Buy    SimpleTradeOrder
	Sell   SimpleTradeOrder
}

func NewSimpleStrategyFromJson(_json []byte) (*SimpleStrategy, error) {
	var s SimpleStrategy

	err := json.Unmarshal(_json, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func SimpleStratedyToJson(s *SimpleStrategy) ([]byte, error) {
	return json.Marshal(&s)
}

func (s SimpleStrategy) Name() string {
	return "Simple"
}

func (s SimpleStrategy) Parameters() []StrategyParameter {
	return []StrategyParameter{
		StrategyParameter{
			Type:  PairParameterType,
			Name:  "Pair",
			Value: s.Pair,
		},
		StrategyParameter{
			Type:  BalanceParameterType,
			Name:  "BaseQuantity",
			Value: Balance{Asset: s.Pair.BaseAsset, Amount: s.BaseQuality},
		},
		StrategyParameter{
			Type:  IntParameterType,
			Name:  "MaxTrades",
			Value: s.MaxTrades,
		},
		StrategyParameter{
			Type:  PercentParameterType,
			Name:  "Profit",
			Value: s.ProfitPercent,
		},
		StrategyParameter{
			Type:  PercentParameterType,
			Name:  "FarPricePercent",
			Value: s.FarPricePercent,
		},
	}
}

func (s SimpleStrategy) ValidateParameter(param StrategyParameter) bool {
	return true
}

func (s *SimpleStrategy) Run(ctx context.Context, _storage interface{}, exchanges []Exchange, logger Logger) error {
	storage, ok := _storage.(SimpleStorage)
	if !ok {
		return errors.New("bad storage type")
	}

	if len(exchanges) != 1 {
		return errors.New("exchanges len != 1")
	}
	exchange := exchanges[0]

	logger.Info("new cycle " + Version)

	balances, err := exchange.Balances([]string{s.Pair.BaseAsset, s.Pair.QuoteAsset})
	if err != nil {
		return fmt.Errorf("balance error: %w", err)
	}

	var baseBalance Balance
	var quoteBalance Balance
	for _, balance := range balances {
		if balance.Asset == s.Pair.BaseAsset {
			baseBalance = balance
		}

		if balance.Asset == s.Pair.QuoteAsset {
			quoteBalance = balance
		}
	}

	logger.Debug(s.Pair.BaseAsset + "= " + baseBalance.Amount.String())
	logger.Debug(s.Pair.QuoteAsset + "= " + quoteBalance.Amount.String())

	openOrders, err := exchange.GetOpenOrders(&OrderFilter{Pairs: []Pair{s.Pair}})
	if err != nil {
		return fmt.Errorf("don't get open orders with error: %w", err)
	}
	logger.Info("open orders: " + strconv.Itoa(len(openOrders)))

	trades, err := storage.GetTrades(&SimpleTradeFilter{Statuses: []SimpleTradeStatus{SimpleTradeStatusBuy, SimpleTradeStatusSell}})
	if err != nil {
		return fmt.Errorf("get trades error: %w", err)
	}

	sellOpenOrders := []SimpleTrade{}
	buyOpenOrders := []SimpleTrade{}
	for _, trade := range trades {
		if trade.Status == SimpleTradeStatusBuy {
			buyOpenOrders = append(buyOpenOrders, trade)
		} else {
			if trade.Sell.OrderId != "" {
				sellOpenOrders = append(sellOpenOrders, trade)
			}
		}
	}

	if len(sellOpenOrders) > 0 || len(buyOpenOrders) > 0 {
		historyOrders, err := exchange.GetHistoryOrders([]Pair{s.Pair})
		if err != nil {
			return fmt.Errorf("get history orders error: %w", err)
		}

		for _, hOrder := range historyOrders {
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			for _, trade := range sellOpenOrders {
				if trade.Sell.OrderId == hOrder.Id {
					if hOrder.Status == FillOrderStatus {
						trade.Status = SimpleTradeStatusFinish
						trade.Sell.Datetime = time.Now().Format(time.RFC3339)

						err := storage.SaveTrade(&trade)
						if err != nil {
							return err
						}
					} else {
						logger.Warn(fmt.Sprintf("trade(id=%d) != FillOrderStatus, %d", trade.Id, hOrder.Status))
					}
					break
				}
			}

			for _, trade := range buyOpenOrders {
				if trade.Buy.OrderId == hOrder.Id {
					if hOrder.Status == FillOrderStatus {
						trade.Status = SimpleTradeStatusSell

						err := storage.SaveTrade(&trade)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	processedTrades := []SimpleTrade{}
	for _, trade := range trades {
		if trade.Status != SimpleTradeStatusFinish {
			processedTrades = append(processedTrades, trade)
		}
	}

	lastPrice, err := exchange.LastPrice(s.Pair)
	if err != nil {
		return fmt.Errorf("Exchange last price error: %w", err)
	}
	logger.Info("current price: " + lastPrice.String())

	amount, isAvailableFunds := s.availableFundCheck(
		quoteBalance.Amount,
		decimal.NewFromFloat(0.0001),
		lastPrice,
		decimal.NewFromInt(10),
		func(amount decimal.Decimal) decimal.Decimal {
			fee, _ := exchange.GetOrderFee(s.Pair, amount, lastPrice)
			return fee.Amount
		},
	)

	var minSpread decimal.Decimal
	for _, trade := range processedTrades {
		farPercent := lastPrice.Sub(trade.Buy.Price).Div(lastPrice).Abs()
		if minSpread.Equal(decimal.Decimal{}) || farPercent.LessThan(minSpread) {
			minSpread = farPercent.RoundUp(5)
		}
	}

	farPrice := minSpread.GreaterThan(s.FarPricePercent)

	logger.Debug(fmt.Sprintf(
		"need new order: max_trades=%t (%d<%d), funds=%t (%s), farPrice=%t (%s>%s)",
		len(processedTrades) < s.MaxTrades,
		len(processedTrades),
		s.MaxTrades,
		isAvailableFunds,
		amount.String(),
		farPrice,
		minSpread.String(),
		s.FarPricePercent.String(),
	))

	if len(processedTrades) < s.MaxTrades && isAvailableFunds && farPrice {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		logger.Info("buy: " + s.BaseQuality.String())

		buyOrder, err := exchange.Buy(s.Pair, amount)
		if err != nil {
			return fmt.Errorf("exchange buy error: %w", err)
		}

		trade := SimpleTrade{
			Status: SimpleTradeStatusBuy,
			Amount: buyOrder.Amount,
			Buy: SimpleTradeOrder{
				OrderId:    buyOrder.Id,
				Datetime:   time.Now().Format(time.RFC3339),
				Price:      buyOrder.Price,
				Commission: buyOrder.Commission,
			},
		}

		if buyOrder.Status == FillOrderStatus {
			trade.Status = SimpleTradeStatusSell
		}

		err = storage.SaveTrade(&trade)
		if err != nil {
			return fmt.Errorf("storage save trades error: %w", err)
		}

		logger.Info(fmt.Sprintf(
			"bought: trade(id=%d), order(id=%s, price=%s)",
			trade.Id,
			trade.Buy.OrderId,
			trade.Buy.Price.String(),
		))

		trades = append(trades, trade)
	}

	for _, trade := range trades {
		if trade.Status == SimpleTradeStatusSell {

			if trade.Sell.OrderId == "" { // sell order didn't created

				sellPrice, err := s.getSellPrice(&trade, &exchange)
				if err != nil {
					logger.Warn(err.Error())
					continue
				}

				select {
				case <-ctx.Done():
					return nil
				default:
				}

				logger.Info(fmt.Sprintf(
					"sell: trade(id=%d), amount=%s, price=%s",
					trade.Id,
					trade.Amount.String(),
					sellPrice.String(),
				))
				sellOrder, err := exchange.Sell(s.Pair, trade.Amount, *sellPrice)
				if err != nil {
					logger.Error(fmt.Sprintf("exchange sell error, trade(id=%d): %#v", trade.Id, err))
					continue
				}

				trade.Sell = SimpleTradeOrder{
					OrderId:    sellOrder.Id,
					Price:      sellOrder.Price,
					Datetime:   time.Now().Format(time.RFC3339),
					Commission: sellOrder.Commission,
				}
				err = storage.SaveTrade(&trade)
				if err != nil {
					return fmt.Errorf("storage save trades error: %w", err)
				}

				logger.Info(fmt.Sprintf(
					"selled: trade(id=%d), order(id=%s, price=%s)",
					trade.Id,
					trade.Sell.OrderId,
					trade.Sell.Price.String(),
				))
			} else {

				sellPrice, err := s.getSellPrice(&trade, &exchange)
				if err != nil {
					logger.Warn(err.Error())
					continue
				}

				if sellPrice.GreaterThan(trade.Sell.Price) {
					logger.Info(fmt.Sprintf(
						"cancelOrder: trade(id=%d), order(id=%s, price=%s), calc price=%s",
						trade.Id,
						trade.Sell.OrderId,
						trade.Sell.Price.String(),
						sellPrice.String(),
					))
					err = exchange.CancelOrder(trade.Sell.OrderId, s.Pair)
					if err != nil {
						logger.Warn(err.Error())
						continue
					}

					trade.Sell.OrderId = ""
					storage.SaveTrade(&trade)
				}

			}
		}
	}

	return nil
}

func (s *SimpleStrategy) availableFundCheck(
	fund decimal.Decimal,
	minAmount decimal.Decimal,
	price decimal.Decimal,
	delta decimal.Decimal,
	getFeeFunc func(amount decimal.Decimal) decimal.Decimal,
) (decimal.Decimal, bool) {
	amount := s.BaseQuality.Copy()
	var isAvailable bool

	for amount.GreaterThanOrEqual(minAmount) {
		isAvailable = amount.GreaterThanOrEqual(minAmount) && fund.GreaterThanOrEqual(amount.Mul(price).Add(getFeeFunc(amount)))
		if isAvailable {
			break
		}
		amount = amount.Div(delta)
	}
	return amount, isAvailable
}

func (s *SimpleStrategy) getSellPrice(trade *SimpleTrade, exchange *Exchange) (*decimal.Decimal, error) {
	paidQuote := trade.Buy.Price.Mul(trade.Amount)

	var paidFeeQuote decimal.Decimal
	if trade.Buy.Commission.Asset == s.Pair.QuoteAsset {
		paidFeeQuote = trade.Buy.Commission.Amount

	} else if trade.Buy.Commission.Asset == s.Pair.BaseAsset {
		paidFeeQuote = trade.Buy.Commission.Amount.Mul(trade.Buy.Price)

	} else {
		return nil, fmt.Errorf("not found commission asset for trade(id=%d)", trade.Id)
	}

	buyPaid := decimal.Sum(paidQuote, paidFeeQuote)

	profit := s.ProfitPercent.Mul(paidQuote)

	//fee, err := (*exchange).GetOrderFee(s.Pair, trade.Amount, trade.Sell.Price)
	fee, err := (*exchange).GetPairFee(s.Pair)
	if err != nil {
		return nil, fmt.Errorf("exchange get order fee error, %w", err)
	}

	sellPrice := decimal.Sum(buyPaid, profit).Div(trade.Amount.Mul(decimal.NewFromInt(1).Sub(fee.Amount))).RoundUp(2)
	return &sellPrice, nil
}
