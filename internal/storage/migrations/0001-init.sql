-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
  id INTEGER NOT NULL PRIMARY KEY,
  links JSON
);

CREATE TABLE IF NOT EXISTS agents (
  id INTEGER NOT NULL PRIMARY KEY,
  user_id INTEGER REFERENCES users,
  status INTEGER NOT NULL,
  strategy_number INTEGER NOT NULL,
  strategy_data JSON
);

CREATE TABLE IF NOT EXISTS exchanges (
  id INTEGER NOT NULL PRIMARY KEY,
  name VARCHAR(16)
);

CREATE TABLE IF NOT EXISTS agent_exchange (
  id INTEGER NOT NULL PRIMARY KEY,
  agent_id INTEGER REFERENCES agents,
  exchange_id INTEGER REFERENCES exchanges
);

CREATE TABLE IF NOT EXISTS user_exchange (
  id INTEGER NOT NULL PRIMARY KEY,
  name VARCHAR(16),
  data JSON,
  user_id INTEGER REFERENCES users,
  exchange_id INTEGER REFERENCES exchanges
);


-- +migrate Down
DROP TABLE users;
DROP TABLE agents;
DROP TABLE exchanges;
DROP TABLE agent_exchange;
DROP TABLE user_exchange;
