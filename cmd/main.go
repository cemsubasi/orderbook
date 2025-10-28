package main

import (
	"context"
	"io"
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
	port := os.Getenv("BE_PORT")
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

	publishers := func() map[string]engine.EventWriter {
		return map[string]engine.EventWriter{
			engine.OrderTopic:    event.NewKafkaPublisher([]string{kafkaBrokers}, engine.OrderTopic),
			engine.TradeTopic:    event.NewKafkaPublisher([]string{kafkaBrokers}, engine.TradeTopic),
			engine.SnapshotTopic: event.NewKafkaPublisher([]string{kafkaBrokers}, engine.SnapshotTopic),
		}
	}()

	// publisher := event.NewKafkaPublisher(
	// 	[]string{kafkaBrokers},
	// 	"orderbook_events",
	// )

	// defer publisher.Close()

	hub := ws.NewWsHub()
	event.StartKafkaWsConsumer(hub, []string{kafkaBrokers})
	event.StartKafkaOrderConsumer([]string{kafkaBrokers}, pgpool)
	event.StartKafkaTradeConsumer([]string{kafkaBrokers}, pgpool)

	books, err := db.RetrieveOrderBooks(pgpool)
	if err != nil {
		log.Fatal("Couldn't load existing orders from DB:", err)
		return
	}

	engine := engine.NewEngine(publishers)
	engine.Setup(books)
	// go engine.Monitor()
	engine.Start(ctx, false)

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	api.HandleOrderController(r, engine)
	ws.HandleEventController(r, engine, hub)

	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HTTP server on %s", ":"+port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
