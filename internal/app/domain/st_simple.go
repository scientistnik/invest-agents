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
	SaveTrade(trade SimpleTrade) error
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
	logger.Debug("Hello in Run SimpleStrategy!" + fmt.Sprintf(" %#v", s))
	if len(exchanges) == 0 {
		return errors.New("exchanges len != 0")
	}
	exchange := exchanges[0]
	storage := _storage.(SimpleStorage)
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

	notFinishTrades := []SimpleTrade{}
	for _, trade := range trades {
		if trade.Sell.OrderId != "" {
			notFinishTrades = append(notFinishTrades, trade)
		}
	}

	if len(notFinishTrades) > 0 {
		historyOrders, err := exchange.GetHistoryOrders([]Pair{s.Pair})
		if err != nil {
			return fmt.Errorf("get history orders error: %w", err)
		}

		for _, trade := range notFinishTrades {
			for _, hOrder := range historyOrders {
				if trade.Sell.OrderId == hOrder.Id {
					if hOrder.Status == FillOrderStatus {
						trade.Status = SimpleTradeStatusFinish
						trade.Sell.Datetime = time.Now().Format(time.RFC3339)
					}
					break
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

	isAvailableFunds := false
	if quoteBalance.Amount.GreaterThan(s.BaseQuality.Mul(lastPrice)) {
		isAvailableFunds = true
	}

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

		buyOrder, err := exchange.Buy(s.Pair, s.BaseQuality)
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

		if buyOrder.Status != FillOrderStatus {
			trade.Status = SimpleTradeStatusSell
		}

		err = storage.SaveTrade(trade)
		if err != nil {
			return fmt.Errorf("storage save trades error: %w", err)
		}

		logger.Info("bought: " + trade.Buy.OrderId + " price=" + trade.Buy.Price.String())
	}

	sellTrades, err := storage.GetTrades(&SimpleTradeFilter{Statuses: []SimpleTradeStatus{SimpleTradeStatusSell}})
	if err != nil {
		return fmt.Errorf("storage get trades error: %w", err)
	}

	for _, trade := range sellTrades {
		if trade.Sell.OrderId == "" {

			fee, err := exchange.GetOrderFee(s.Pair, trade.Amount)
			if err != nil {
				return fmt.Errorf("exchange get order fee error, %w", err)
			}

			quantity := trade.Amount

			paidQuote := trade.Buy.Price.Mul(trade.Amount)

			var paidFeeQuote decimal.Decimal
			if trade.Buy.Commission.Asset == s.Pair.QuoteAsset {
				paidFeeQuote = trade.Buy.Commission.Amount

			} else if trade.Buy.Commission.Asset == s.Pair.BaseAsset {
				paidFeeQuote = trade.Buy.Commission.Amount.Mul(trade.Buy.Price)

			} else {
				continue
			}

			buyPaid := decimal.Sum(paidQuote, paidFeeQuote)

			profit := s.ProfitPercent.Mul(trade.Buy.Price).Mul(trade.Amount)
			futureFee := fee

			sellPrice := decimal.Sum(buyPaid, profit, *futureFee).Div(trade.Amount)

			logger.Info("sell: " + trade.Amount.String())
			sellOrder, err := exchange.Sell(s.Pair, quantity, sellPrice)
			if err != nil {
				return fmt.Errorf("exchange sell error: %w", err)
			}

			trade.Sell = SimpleTradeOrder{
				OrderId:    sellOrder.Id,
				Price:      sellOrder.Price,
				Datetime:   time.Now().Format(time.RFC3339),
				Commission: sellOrder.Commission,
			}
			err = storage.SaveTrade(trade)
			if err != nil {
				return fmt.Errorf("storage save trades error: %w", err)
			}

			logger.Info("selled: id=" + trade.Sell.OrderId)
		}
	}

	return nil
}
