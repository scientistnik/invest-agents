package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/scientistnik/invest-agents/internal/app"
	"github.com/scientistnik/invest-agents/internal/app/domain"
	"log"

	_ "github.com/mattn/go-sqlite3"
	migrate "github.com/rubenv/sql-migrate"
)

func GetSqliteAppStorage(filename string) (*AppStorage, error) {
	driver := SqliteDriver{filename: filename}
	err := driver.migrateRollUp()
	if err != nil {
		return nil, err
	}

	return &AppStorage{driver: &driver}, nil
}

type SqliteDriver struct {
	filename string
	db       *sql.DB
}

var _ Driver = (*SqliteDriver)(nil)

func (s *SqliteDriver) connect() error {
	db, err := sql.Open("sqlite3", s.filename)
	if err != nil {
		return err
	}

	s.db = db

	return nil
}

func (s SqliteDriver) disconnect() error {
	return s.db.Close()
}

func (s SqliteDriver) migrateRollUp() error {
	db, err := sql.Open("sqlite3", s.filename)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	n, err := migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		return err
	}

	if n > 0 {
		fmt.Printf("Applied %d migrations!\n", n)
	}

	return nil
}

func (s SqliteDriver) migrateRollDown() error {
	return nil
}

func (s SqliteDriver) getDB() *sql.DB {
	return s.db
}

func (s SqliteDriver) userGetOrCreate(links app.UserLinks) (*domain.User, error) {
	rows, err := s.db.Query("SELECT id from users where json_extract(links, '$.telegram') = ?", links.Telegram)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user := domain.User{}
	userCount := 0
	for rows.Next() {
		if userCount != 0 {
			return nil, errors.New("found more than one User")
		}

		err = rows.Scan(&user.Id)
		if err != nil {
			return nil, fmt.Errorf("error in get user (scan row): %w", err)
		}

		userCount++
	}

	if userCount == 1 {
		return &user, nil
	}

	jsonLinks, err := json.Marshal(links)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec("INSERT INTO users (links) values (?)", jsonLinks)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &domain.User{Id: id}, nil
}

func (s SqliteDriver) agentFind(filter app.AgentFilter) ([]domain.Agent, error) {
	agents := []domain.Agent{}

	query := BaseSelectAgensQuery

	predicats := []string{}
	queryArgs := []interface{}{}

	if filter.Id != 0 {
		predicats = append(predicats, "(id=?)")
		queryArgs = append(queryArgs, filter.Id)
	}

	if filter.Status != domain.ErrorAgentStatus {
		predicats = append(predicats, "(status=?)")
		queryArgs = append(queryArgs, filter.Status)
	}

	if filter.UserId != 0 {
		predicats = append(predicats, "(user_id=?)")
		queryArgs = append(queryArgs, filter.UserId)
	}

	if len(predicats) > 0 {
		query += " where "
		for index, predicat := range predicats {
			if index != 0 {
				query += " and "
			}
			query += predicat
		}
	}

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("error in agentFind (query): %w", err)
	}

	for rows.Next() {
		agent := domain.Agent{}
		err = rows.Scan(&agent.Id, &agent.UserId, &agent.Status, &agent.StrategyId, &agent.StrategyData)
		if err != nil {
			return nil, fmt.Errorf("error in agentFind (scan row): %w", err)
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

func (s SqliteDriver) agentCreate(agent domain.Agent) (*domain.Agent, error) {
	result, err := s.db.Exec(
		"INSERT INTO agents (user_id, status, strategy_number, strategy_data) values (?,?,?,?)",
		agent.UserId,
		agent.Status,
		agent.StrategyId,
		agent.StrategyData,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	agent.Id = id
	return &agent, nil
}

func (s SqliteDriver) agentSetStatus(agent *domain.Agent, status domain.AgentStatus) error {
	agent.Status = status

	_, err := s.db.Exec("UPDATE agents set status=? where id=?", agent.Status, agent.Id)
	return err
}

func (s SqliteDriver) agentUpdateData(agent *domain.Agent, data []byte) error {
	agent.StrategyData = data

	_, err := s.db.Exec("UPDATE agents set strategy_data=? where id=?", agent.StrategyData, agent.Id)
	return err
}

func (s SqliteDriver) getAgentExchanges(agentId int64) ([]app.ExchangeData, error) {
	exchanges := []app.ExchangeData{}

	rows, err := s.db.Query(SelectAgentExchangesQuery, agentId)
	if err != nil {
		return nil, fmt.Errorf("error in getAgentExchanges (query): %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		exchange := app.ExchangeData{}
		err = rows.Scan(&exchange.Id, &exchange.Data)
		if err != nil {
			return nil, fmt.Errorf("error in getAgentExchanges (scan row): %w", err)
		}

		exchanges = append(exchanges, exchange)
	}

	return exchanges, nil
}

func (s SqliteDriver) findExchanges(filter app.ExchangeFilter) ([]app.ExchangeData, error) {
	exchanges := []app.ExchangeData{}

	query := SelectUserExchangesQuery

	predicats := []string{}
	queryArgs := []interface{}{}

	if filter.UserId != 0 {
		predicats = append(predicats, "(user_id=?)")
		queryArgs = append(queryArgs, filter.UserId)
	}

	if filter.ExchangeNumber != 0 {
		predicats = append(predicats, "(exchange_number=?)")
		queryArgs = append(queryArgs, filter.ExchangeNumber)
	}

	if len(predicats) > 0 {
		query += " where "
		for index, predicat := range predicats {
			if index != 0 {
				query += " and "
			}
			query += predicat
		}
	}

	rows, err := s.db.Query(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("error in filterExchanges (query): %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		exchange := app.ExchangeData{}
		err = rows.Scan(&exchange.Id, &exchange.Data)
		if err != nil {
			return nil, fmt.Errorf("error in getAgentExchanges (scan row): %w", err)
		}

		exchanges = append(exchanges, exchange)
	}

	return exchanges, nil
}

func (s SqliteDriver) addExchange(userId int64, exchangeNumber int, data []byte) error {
	_, err := s.db.Exec(
		"INSERT INTO exchanges (user_id, exchange_number, data) values (?,?,?)",
		userId,
		exchangeNumber,
		data,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s SqliteDriver) agentAddExchange(agent *domain.Agent, exchanges []app.ExchangeData) error {
	for _, exchange := range exchanges {
		_, err := s.db.Exec(
			"INSERT INTO agent_exchange (agent_id, exchange_id) values (?,?)",
			agent.Id,
			exchange.Id,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
