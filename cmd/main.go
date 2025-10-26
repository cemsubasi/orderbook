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
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")

	if pgUser == "" || pgPass == "" {
		log.Println("Environment variables not set.")
		return
	}
	if pgHost == "" {
		pgHost = "localhost"
	}
	if kafkaHost == "" {
		kafkaHost = "localhost"
	}
	if kafkaPort == "" {
		kafkaPort = "9092"
	}

	kafkaBrokers := kafkaHost + ":" + kafkaPort

	pgpool := db.InitPostgres(pgUser, pgPass, pgHost, pgDB)
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

	log.Printf("Starting HTTP server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
