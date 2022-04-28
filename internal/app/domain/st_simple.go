package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type SimpleStrategy struct {
	pair            Pair
	baseQuality     decimal.Decimal
	maxTrades       int
	profitPercent   decimal.Decimal
	farPricePercent decimal.Decimal
}

type TradeStatus = int

const (
	_                             = iota
	TradeStatusBuy    TradeStatus = iota
	TradeStatusSell   TradeStatus = iota
	TradeStatusFinish TradeStatus = iota
)

type SimpleStorage interface {
	//Storage
	GetTrades([]TradeStatus) ([]Trade, error)
	SaveTrades(trades []Trade) error
}

type TradeOrder struct {
	datetime   string
	orderId    string
	price      decimal.Decimal
	amount     decimal.Decimal
	commission Balance
}

type Trade struct {
	status TradeStatus
	buy    *TradeOrder
	sell   *TradeOrder
}

func ParseJSONToSimpleStrategy(_json []byte) (*SimpleStrategy, error) {
	var s SimpleStrategy

	err := json.Unmarshal(_json, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func NewSimpleStrategyFromJson(_json []byte) *SimpleStrategy {
	return nil
}

func (s *SimpleStrategy) Run(_storage interface{}, exchanges []Exchange, logger Logger) error {
	logger.Info("Hello in Run SimpleStrategy!")
	if len(exchanges) == 0 {
		logger.Info("Exchanges len = 0!")
		return nil
	}
	exchange := exchanges[0]
	storage := _storage.(SimpleStorage)
	logger.Info("new cycle" + time.Now().Format(time.RFC3339))

	balances, err := exchange.Balances([]string{s.pair.BaseAsset, s.pair.QuoteAsset})
	if err != nil {
		return fmt.Errorf("balance error: %w", err)
	}

	var baseBalance Balance
	var quoteBalance Balance
	for _, balance := range balances {
		if balance.Asset == s.pair.BaseAsset {
			baseBalance = balance
		}

		if balance.Asset == s.pair.QuoteAsset {
			quoteBalance = balance
		}
	}

	logger.Debug(s.pair.BaseAsset + "= " + baseBalance.Amount.String())
	logger.Debug(s.pair.QuoteAsset + "= " + quoteBalance.Amount.String())

	openOrders, err := exchange.GetOpenOrders()
	if err != nil {
		return fmt.Errorf("don't get open orders with error: %w", err)
	}
	logger.Info("open orders: " + strconv.Itoa(len(openOrders)))

	trades, err := storage.GetTrades([]TradeStatus{TradeStatusBuy, TradeStatusSell})
	if err != nil {
		return fmt.Errorf("get trades error: %w", err)
	}

	notFinishTrades := []Trade{}
	for _, trade := range trades {
		if trade.sell != nil {
			notFinishTrades = append(notFinishTrades, trade)
		}
	}

	if len(notFinishTrades) > 0 {
		historyOrders, err := exchange.GetHistoryOrders([]Pair{s.pair})
		if err != nil {
			return fmt.Errorf("get history orders error: %w", err)
		}

		for _, trade := range notFinishTrades {
			for _, hOrder := range historyOrders {
				if trade.sell.orderId == hOrder.Id {
					if hOrder.Status == OrderStatusFill {
						trade.status = TradeStatusFinish
						trade.sell.datetime = time.Now().Format(time.RFC3339)
					}
					break
				}
			}
		}
	}

	processedTrades := []Trade{}
	for _, trade := range trades {
		if trade.status != TradeStatusFinish {
			processedTrades = append(processedTrades, trade)
		}
	}

	lastPrice, err := exchange.LastPrice(s.pair)
	if err != nil {
		return fmt.Errorf("Exchange last price error: %w", err)
	}
	logger.Info("current price: " + lastPrice.String())

	isAvailableFunds := false
	if quoteBalance.Amount.GreaterThan(s.baseQuality.Mul(lastPrice)) {
		isAvailableFunds = true
	}

	farPrice := true
	for _, trade := range processedTrades {
		farPercent := lastPrice.Sub(trade.buy.price).Div(lastPrice).Abs()
		if farPercent.GreaterThan(s.farPricePercent) {
			farPrice = false
			break
		}
	}

	logger.Debug("need new order: max_trades=" + strconv.FormatBool(
		len(processedTrades) < s.maxTrades) + ", funds=" + strconv.FormatBool(isAvailableFunds) + ", farPrice=" + strconv.FormatBool(farPrice))

	if len(processedTrades) < s.maxTrades && isAvailableFunds && farPrice {
		logger.Info("buy: " + s.baseQuality.String())

		buyOrder, err := exchange.Buy(s.pair, s.baseQuality)
		if err != nil {
			return fmt.Errorf("exchange buy error: %w", err)
		}

		trade := Trade{
			status: TradeStatusBuy,
			buy: &TradeOrder{
				orderId:    buyOrder.Id,
				datetime:   time.Now().Format(time.RFC3339),
				price:      buyOrder.Price,
				amount:     buyOrder.Amount,
				commission: buyOrder.Commission,
			},
		}

		if buyOrder.Status != OrderStatusFill {
			trade.status = TradeStatusSell
		}

		err = storage.SaveTrades([]Trade{trade})
		if err != nil {
			return fmt.Errorf("storage save trades error: %w", err)
		}

		logger.Info("bought: " + trade.buy.orderId + " price=" + trade.buy.price.String())
	}

	sellTrades, err := storage.GetTrades([]TradeStatus{TradeStatusSell})
	if err != nil {
		return fmt.Errorf("storage get trades error: %w", err)
	}

	for _, trade := range sellTrades {
		if trade.sell == nil {

			fee, err := exchange.GetOrderFee(s.pair, trade.buy.amount)
			if err != nil {
				return fmt.Errorf("exchange get order fee error, %w", err)
			}

			quantity := trade.buy.amount

			paidQuote := trade.buy.price.Mul(trade.buy.amount)

			var paidFeeQuote decimal.Decimal
			if trade.buy.commission.Asset == s.pair.QuoteAsset {
				paidFeeQuote = trade.buy.commission.Amount

			} else if trade.buy.commission.Asset == s.pair.BaseAsset {
				paidFeeQuote = trade.buy.commission.Amount.Mul(trade.buy.price)

			} else {
				continue
			}

			buyPaid := decimal.Sum(paidQuote, paidFeeQuote)

			profit := s.profitPercent.Mul(trade.buy.price).Mul(trade.buy.amount)
			futureFee := fee

			sellPrice := decimal.Sum(buyPaid, profit, *futureFee).Div(trade.buy.amount)

			logger.Info("sell: " + trade.buy.amount.String())
			sellOrder, err := exchange.Sell(s.pair, quantity, sellPrice)
			if err != nil {
				return fmt.Errorf("exchange sell error: %w", err)
			}

			trade.sell = &TradeOrder{
				orderId:    sellOrder.Id,
				amount:     sellOrder.Amount,
				price:      sellOrder.Price,
				datetime:   time.Now().Format(time.RFC3339),
				commission: sellOrder.Commission,
			}
			err = storage.SaveTrades([]Trade{trade})
			if err != nil {
				return fmt.Errorf("storage save trades error: %w", err)
			}

			logger.Info("selled: id=" + trade.sell.orderId)
		}
	}

	return nil
}
