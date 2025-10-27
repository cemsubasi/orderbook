#!/bin/sh
migrate -path internal/db/migrations -database "postgres://${PG_USER}:${PG_PASS}@${PG_HOST}:5432/${PG_DB}?sslmode=disable" down $1

