package main

import (
	"context"
	"log"
	"os"

	"github.com/cemsubasi/orderbook/internal/api"
	"github.com/cemsubasi/orderbook/internal/db"
	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/gin-gonic/gin"
)

func main() {
	pgUser := os.Getenv("PG_USER")
	pgPass := os.Getenv("PG_PASS")
	kafkaUser := os.Getenv("KAFKA_USER")
	kafkaPass := os.Getenv("KAFKA_PASS")
	port := os.Getenv("OB_PORT")

	if pgUser == "" || pgPass == "" || kafkaUser == "" || kafkaPass == "" {
		log.Println("Environment variables not set.")
		return
	}

	pgpool := db.InitPostgres(pgUser, pgPass)
	if pgpool == nil {
		log.Fatal("Posgres couldn't initialized.")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine := engine.NewEngine()
	engine.Start(ctx)

	r := gin.Default()

	api.HandleOrderController(r, engine)

	addr := ":8080"
	if port != "" {
		addr = ":" + port
	}

	log.Printf("Starting HTTP server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
