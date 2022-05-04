package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type SimpleStrategy struct {
	Pair            Pair            `json:"pair"`
	BaseQuality     decimal.Decimal `json:"base_quality"`
	MaxTrades       int             `json:"max_trades"`
	ProfitPercent   decimal.Decimal `json:"profit_percent"`
	FarPricePercent decimal.Decimal `json:"far_price_percent"`
}

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

func (s *SimpleStrategy) Run(_storage interface{}, exchanges []Exchange, logger Logger) error {
	storage, ok := _storage.(SimpleStorage)
	if !ok {
		return errors.New("bad storage type")
	}

	if len(exchanges) != 1 {
		return errors.New("exchanges len != 1")
	}
	exchange := exchanges[0]

	logger.Info("new cycle " + time.Now().Format(time.RFC3339))

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

	farPrice := true
	for _, trade := range processedTrades {
		farPercent := lastPrice.Sub(trade.Buy.Price).Div(lastPrice).Abs()
		if farPercent.GreaterThan(s.FarPricePercent) {
			farPrice = false
			break
		}
	}

	logger.Debug("need new order: max_trades=" + strconv.FormatBool(
		len(processedTrades) < s.MaxTrades) + ", funds=" + strconv.FormatBool(isAvailableFunds) + ", farPrice=" + strconv.FormatBool(farPrice))

	if len(processedTrades) < s.MaxTrades && isAvailableFunds && farPrice {
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

		logger.Info("bought: " + trade.Buy.OrderId + " price=" + trade.Buy.Price.String())

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

				logger.Info("sell: " + trade.Amount.String())
				sellOrder, err := exchange.Sell(s.Pair, trade.Amount, *sellPrice)
				if err != nil {
					return fmt.Errorf("exchange sell error: %w", err)
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

				logger.Info("selled: id=" + trade.Sell.OrderId)

			} else {

				sellPrice, err := s.getSellPrice(&trade, &exchange)
				if err != nil {
					logger.Warn(err.Error())
					continue
				}

				if trade.Sell.Price != *sellPrice {
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
	fee, err := (*exchange).GetOrderFee(s.Pair, trade.Amount, trade.Sell.Price)
	if err != nil {
		return nil, fmt.Errorf("exchange get order fee error, %w", err)
	}

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

	profit := s.ProfitPercent.Mul(trade.Buy.Price).Mul(trade.Amount)
	futureFee := fee.Amount

	sellPrice := decimal.Sum(buyPaid, profit, futureFee).Div(trade.Amount)
	return &sellPrice, nil
}
