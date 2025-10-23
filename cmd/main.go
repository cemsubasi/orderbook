package main

import (
	"fmt"
	"os"
)

func main () {
	pgUser := os.Getenv("PG_USER")
	pgPass := os.Getenv("PG_PASS")
	kafkaUser := os.Getenv("KAFKA_USER")
	kafkaPass := os.Getenv("KAFKA_PASS")

	if pgUser == "" || pgPass == "" || kafkaUser == "" || kafkaPass == "" {
		fmt.Println("Environment variables not set")
		return
	}
}
