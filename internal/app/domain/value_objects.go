package domain

import "github.com/shopspring/decimal"

type Balance struct {
	Asset  string          `json:"asset"`
	Amount decimal.Decimal `json:"amount"`
}

type Pair struct {
	BaseAsset  string `json:"base_asset"`
	QuoteAsset string `json:"quote_asset"`
}
