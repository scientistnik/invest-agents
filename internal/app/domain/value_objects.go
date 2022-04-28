package domain

import "github.com/shopspring/decimal"

type Balance struct {
	Asset  string
	Amount decimal.Decimal
}

type Pair struct {
	BaseAsset  string
	QuoteAsset string
}
