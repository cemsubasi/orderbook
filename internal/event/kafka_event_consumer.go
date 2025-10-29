package event

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func StartKafkaOrderConsumer(brokers []string, db *pgxpool.Pool, ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				reader := kafka.NewReader(kafka.ReaderConfig{
					Brokers: brokers,
					Topic:   engine.OrderTopic,
					GroupID: "order_handler",
				})
				log.Println("Kafka order consumer connected")
				for {
					select {
					case <-ctx.Done():
						log.Println("Kafka consumer stopping during read...")
						_ = reader.Close()
						return
					default:
						m, err := reader.ReadMessage(ctx)
						if err != nil {
							if errors.Is(err, context.Canceled) {
								log.Println("Kafka consumer read canceled due to context shutdown")
								return
							}

							log.Println("kafka read err:", err)
							log.Println("Reconnecting to Kafka in 1 second...")
							_ = reader.Close()
							time.Sleep(1 * time.Second)
							break
						}

						var event struct {
							Type    string          `json:"type"`
							Payload json.RawMessage `json:"payload"`
						}
						if err := json.Unmarshal(m.Value, &event); err != nil {
							log.Fatal("unmarshal event err:", err)
							continue
						}

						if event.Type != "order_added" {
							log.Fatal("The kafka message is in the wrong topic")
							return
						}

						var order *engine.Order
						if err := json.Unmarshal(event.Payload, &order); err != nil {
							log.Fatal("unmarshal order err:", err)
							continue
						}

						persistOrder(ctx, db, order)
					}
				}
			}
		}
	}()
}

func StartKafkaTradeConsumer(brokers []string, db *pgxpool.Pool, ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				log.Println("Creating reader...")
				reader := kafka.NewReader(kafka.ReaderConfig{
					Brokers: brokers,
					Topic:   engine.TradeTopic,
					GroupID: "trade_handler",
				})
				log.Println("Kafka trade consumer connected")

				for {
					select {
					case <-ctx.Done():
						log.Println("Kafka consumer stopping during read...")
						_ = reader.Close()
						return
					default:
						m, err := reader.ReadMessage(ctx)
						if err != nil {
							if errors.Is(err, context.Canceled) {
								log.Println("Kafka consumer read canceled due to context shutdown")
								return
							}

							log.Println("kafka read err:", err)
							log.Println("Reconnecting to Kafka in 1 second...")
							_ = reader.Close()
							time.Sleep(1 * time.Second)
							break
						}

						var event struct {
							Type    string          `json:"type"`
							Payload json.RawMessage `json:"payload"`
						}
						if err := json.Unmarshal(m.Value, &event); err != nil {
							log.Println("unmarshal event err:", err)
							continue
						}

						if event.Type != "order_matched" {
							log.Fatal("The kafka message is in the wrong topic")
							continue
						}

						var trades []*engine.Trade
						if err := json.Unmarshal(event.Payload, &trades); err != nil {
							log.Println("unmarshal trade err:", err)
							continue
						}

						persistTrades(ctx, db, trades)
					}
				}
			}
		}
	}()
}

func persistOrder(ctx context.Context, db *pgxpool.Pool, order *engine.Order) {
	_, err := db.Exec(ctx, `INSERT INTO orders (id, symbol, side, price, quantity, remaining, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		order.ID, order.Symbol, order.Side, order.Price, order.Quantity, order.Remaining, time.Now().UTC())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Database operation canceled due to context shutdown")
			return
		}

		log.Fatal("persist order err:", err)
	} else {
		// log.Printf("Order %s persisted/updated with remaining %f", order.ID, order.Remaining)
	}
}

func persistTrades(ctx context.Context, db *pgxpool.Pool, trades []*engine.Trade) {
	columns := []string{"id", "symbol", "buy_order_id", "sell_order_id", "price", "quantity", "executed_at"}

	rows := make([][]interface{}, len(trades))
	for i, t := range trades {
		rows[i] = []interface{}{t.ID, t.Symbol, t.BuyOrderID, t.SellOrderID, t.Price, t.Quantity, t.ExecutedAt}
	}

	_, err := db.CopyFrom(
		ctx,
		pgx.Identifier{"trades"},
		columns,
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Database operation canceled due to context shutdown")
			return
		}

		log.Fatal("persist trades err:", err)
	}
}
