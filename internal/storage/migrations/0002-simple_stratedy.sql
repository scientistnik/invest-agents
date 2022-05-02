-- +migrate Up
CREATE TABLE IF NOT EXISTS st_simple_trades (
  id INTEGER NOT NULL PRIMARY KEY,
  agent_id INTEGER REFERENCES agents,
  status INTEGER NOT NULL,
  buy_order_id VARCHAR(256),
  buy_datetime INTEGER,
  buy_price VARCHAR(16),
  buy_amount VARCHAR(16),
  buy_commission VARCHAR(16)
);

-- +migrate Down
DROP TABLE st_simple_trades;