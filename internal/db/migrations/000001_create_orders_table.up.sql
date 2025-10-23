CREATE TABLE IF NOT EXISTS orders (
	id TEXT PRIMARY KEY,
	symbol TEXT,
	side TEXT,
	price NUMERIC,
	quantity NUMERIC,
	remaining NUMERIC,
	status TEXT,
	created_at TIMESTAMP
);
