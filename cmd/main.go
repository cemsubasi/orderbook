package main

import (
	"log"
	"os"

	"github.com/cemsubasi/orderbook/internal/db"
)

func main () {
	pgUser := os.Getenv("PG_USER")
	pgPass := os.Getenv("PG_PASS")
	kafkaUser := os.Getenv("KAFKA_USER")
	kafkaPass := os.Getenv("KAFKA_PASS")

	if pgUser == "" || pgPass == "" || kafkaUser == "" || kafkaPass == "" {
		log.Println("Environment variables not set.")
		return
	}

    pgpool := db.InitPostgres(pgUser, pgPass)
		if pgpool == nil {
			log.Fatal("Posgres couldn't initialized.")
			return
		}

		log.Println("Process is closing successfuly.")
}
