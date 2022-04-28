package storage

const UserInsertQuery = "INSERT INTO users (id) values (1)"

const BaseSelectAgensQuery = "SELECT id, user_id, status, strategy_number, strategy_data FROM agents"

const SelectAgentExchangesQuery = `
SELECT 
	ae.exchange_id as id,
	ue.data as data
FROM agents as a 
JOIN agent_exchange as ae 
	ON a.id = ae.agent_id 
JOIN user_exchange as ue
	ON a.user_id = ue.user_id
WHERE a.id = ?
`
