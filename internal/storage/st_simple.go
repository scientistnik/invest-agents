package storage

import (
	"database/sql"
	"fmt"
	"github.com/scientistnik/invest-agents/internal/app/domain"
	"github.com/shopspring/decimal"
	"strings"
)

type SimpleStorage struct {
	agent domain.Agent
	db    *sql.DB
}

var _ domain.SimpleStorage = (*SimpleStorage)(nil)

func (ss SimpleStorage) GetTrades(filter *domain.SimpleTradeFilter) ([]domain.SimpleTrade, error) {
	query := `
	SELECT 
		id, 
		status,
		amount,
		buy_order_id,
		buy_datetime,
		buy_price,
		buy_commission,
		buy_commission_asset,
		sell_order_id,
		sell_datetime,
		sell_price,
		sell_commission,
		sell_commission_asset 
	FROM st_simple_trades
	WHERE `

	predicats := []string{"(agent_id=?)"}
	queryArgs := []interface{}{ss.agent.Id}

	if filter != nil {

		if len(filter.Statuses) > 0 {
			questions := []string{}
			for _, status := range filter.Statuses {
				questions = append(questions, "?")
				queryArgs = append(queryArgs, status)
			}

			predicats = append(predicats, " (status in ("+strings.Join(questions, ",")+"))")
		}
	}

	query += strings.Join(predicats, " and ")

	rows, err := ss.db.Query(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("error in GetTrades (query): %w", err)
	}
	defer rows.Close()

	trades := []domain.SimpleTrade{}
	for rows.Next() {
		trade := domain.SimpleTrade{}

		result := struct {
			sellOrderId          sql.NullString
			sellDatetime         sql.NullString
			sellPrice            sql.NullString
			sellCommissionAmount sql.NullString
			sellCommissionAsset  sql.NullString
		}{}

		err = rows.Scan(
			&trade.Id,
			&trade.Status,
			&trade.Amount,
			&trade.Buy.OrderId,
			&trade.Buy.Datetime,
			&trade.Buy.Price,
			&trade.Buy.Commission.Amount,
			&trade.Buy.Commission.Asset,
			&result.sellOrderId,
			&result.sellDatetime,
			&result.sellPrice,
			&result.sellCommissionAmount,
			&result.sellCommissionAsset,
		)

		trade.Sell.OrderId = result.sellOrderId.String
		trade.Sell.Datetime = result.sellDatetime.String
		trade.Sell.Price, _ = decimal.NewFromString(result.sellPrice.String)
		trade.Sell.Commission.Amount, _ = decimal.NewFromString(result.sellCommissionAmount.String)
		trade.Sell.Commission.Asset = result.sellCommissionAsset.String

		if err != nil {
			return nil, fmt.Errorf("error in GetTrades (scan row): %w", err)
		}

		trades = append(trades, trade)
	}

	return trades, nil
}

func (ss SimpleStorage) SaveTrade(trade *domain.SimpleTrade) error {
	var err error
	if trade.Id == 0 {
		_, err = ss.db.Exec(`
		INSERT INTO st_simple_trades (
			agent_id,
			status,
			amount,
			buy_order_id,
			buy_datetime,
			buy_price,
			buy_commission,
			buy_commission_asset,
			sell_order_id,
			sell_datetime,
			sell_price,
			sell_commission,
			sell_commission_asset
		)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			ss.agent.Id,
			trade.Status,
			trade.Amount,
			trade.Buy.OrderId,
			trade.Buy.Datetime,
			trade.Buy.Price,
			trade.Buy.Commission.Amount,
			trade.Buy.Commission.Asset,
			trade.Sell.OrderId,
			trade.Sell.Datetime,
			trade.Sell.Price,
			trade.Sell.Commission.Amount,
			trade.Sell.Commission.Asset,
		)
	} else {
		_, err = ss.db.Exec(`
		UPDATE st_simple_trades set
			agent_id=?,
			status=?,
			amount=?,
			buy_order_id=?,
			buy_datetime=?,
			buy_price=?,
			buy_commission=?,
			buy_commission_asset=?,
			sell_order_id=?,
			sell_datetime=?,
			sell_price=?,
			sell_commission=?,
			sell_commission_asset=?
		WHERE id=?`,
			ss.agent.Id,
			trade.Status,
			trade.Amount,
			trade.Buy.OrderId,
			trade.Buy.Datetime,
			trade.Buy.Price,
			trade.Buy.Commission.Amount,
			trade.Buy.Commission.Asset,
			trade.Sell.OrderId,
			trade.Sell.Datetime,
			trade.Sell.Price,
			trade.Sell.Commission.Amount,
			trade.Sell.Commission.Asset,
			trade.Id,
		)
	}
	if err != nil {
		return err
	}

	return nil
}
