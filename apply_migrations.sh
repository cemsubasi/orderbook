#!/bin/bash
migrate -path internal/db/migrations -database "postgres://${PG_USER}:${PG_PASS}@localhost:5432/orderbook?sslmode=disable" up

