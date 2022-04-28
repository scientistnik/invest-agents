package storage

import "github.com/scientistnik/invest-agents/internal/app/domain"

type SimpleStorage struct {
	agent domain.Agent
}

func (ss SimpleStorage) GetTrades([]domain.TradeStatus) ([]domain.Trade, error) {
	return nil, nil
}

func (ss SimpleStorage) SaveTrades(trades []domain.Trade) error {
	return nil
}
