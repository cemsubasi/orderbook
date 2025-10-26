package main

import (
	"context"
	"log"
	"os"

	"github.com/cemsubasi/orderbook/internal/api"
	"github.com/cemsubasi/orderbook/internal/db"
	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/cemsubasi/orderbook/internal/event"
	"github.com/cemsubasi/orderbook/internal/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	pgUser := os.Getenv("PG_USER")
	pgPass := os.Getenv("PG_PASS")
	pgDB := os.Getenv("PG_DB")
	pgHost := os.Getenv("PG_HOST")
	pgPort := os.Getenv("PG_PORT")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	port := os.Getenv("OB_PORT")

	if pgUser == "" || pgPass == "" {
		log.Println("Environment variables not set.")
		return
	}

	if pgHost == "" {
		pgHost = "localhost"
	}

	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	pgpool := db.InitPostgres(pgUser, pgPass, pgHost, pgPort, pgDB)
	if pgpool == nil {
		log.Fatal("Posgres couldn't initialized.")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := ws.NewWsHub()
	ws.StartKafkaToWsWorker(hub, []string{kafkaBrokers}, "orderbook_events")

	event.StartOrderEventKafkaConsumer([]string{kafkaBrokers}, "orderbook_events", pgpool)
	event.StartTradeEventKafkaConsumer([]string{kafkaBrokers}, "orderbook_events", pgpool)

	publisher := event.NewKafkaPublisher(
		[]string{kafkaBrokers},
		"orderbook_events",
	)

	defer publisher.Close()

	books, err := db.LoadOrderBooks(pgpool)
	if err != nil {
		log.Fatal("Couldn't load existing orders from DB:", err)
		return
	}

	engine := engine.NewEngine(publisher)
	engine.Setup(books)
	engine.Start(ctx)

	r := gin.Default()
	r.Use(cors.Default())

	api.HandleOrderController(r, engine)
	ws.HandleEventController(r, engine, hub)

	addr := ":8080"
	if port != "" {
		addr = ":" + port
	}

	log.Printf("Starting HTTP server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
