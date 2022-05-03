-- +migrate Up
CREATE TABLE IF NOT EXISTS st_simple_trades (
  id INTEGER NOT NULL PRIMARY KEY,
  agent_id INTEGER REFERENCES agents,
  status INTEGER NOT NULL,
  amount VARCHAR(16),
  buy_order_id VARCHAR(256),
  buy_datetime VARCHAR(16),
  buy_price VARCHAR(16),
  buy_commission VARCHAR(16),
  buy_commission_asset VARCHAR(16),
  sell_order_id VARCHAR(256),
  sell_datetime VARCHAR(16),
  sell_price VARCHAR(16),
  sell_commission VARCHAR(16),
  sell_commission_asset VARCHAR(16)
);

-- +migrate Down
DROP TABLE st_simple_trades;