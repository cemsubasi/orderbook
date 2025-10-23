CREATE TABLE IF NOT EXISTS trades (
	id TEXT PRIMARY KEY,
	symbol TEXT,
	buy_order_id TEXT,
	sell_order_id TEXT,
	price NUMERIC,
	quantity NUMERIC,
	executed_at TIMESTAMP
);
