package storage

import (
	"errors"
	"github.com/scientistnik/invest-agents/internal/app/domain"
)

type SimpleStorage struct {
	agent domain.Agent
}

func (ss SimpleStorage) GetTrades([]domain.TradeStatus) ([]domain.Trade, error) {
	return nil, errors.New("not implemented")
}

func (ss SimpleStorage) SaveTrades(trades []domain.Trade) error {
	return errors.New("not implemented")
}
