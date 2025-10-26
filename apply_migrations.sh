#!/bin/sh
migrate -path internal/db/migrations -database "postgres://${PG_USER}:${PG_PASS}@${PG_HOST}:${OB_PORT}/${PG_DB}?sslmode=disable" up

