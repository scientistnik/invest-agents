package storage

const UserInsertQuery = "INSERT INTO users (id) values (1)"

const BaseSelectAgensQuery = "SELECT id, user_id, status, strategy_number, strategy_data FROM agents"

const SelectAgentExchangesQuery = `
SELECT 
	ue.exchange_number as id,
	ue.data as data
FROM agents as a 
JOIN agent_exchange as ae 
	ON a.id = ae.agent_id 
JOIN exchanges as ue
	ON ae.exchange_id = ue.id
WHERE a.id = ?
`

const SelectUserExchangesQuery = "SELECT id, data from exchanges"
