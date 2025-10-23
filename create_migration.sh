#!/bin/bash

migrate create -ext sql -dir internal/db/migrations -seq $1

