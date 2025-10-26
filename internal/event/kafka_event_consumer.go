package event

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func StartOrderEventKafkaConsumer(brokers []string, topic string, db *pgxpool.Pool) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "order_handler",
	})

	go func() {
		ctx := context.Background()

		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Println("kafka read err:", err)
				continue
			}

			var event struct {
				Type    string          `json:"type"`
				Payload json.RawMessage `json:"payload"`
			}
			if err := json.Unmarshal(m.Value, &event); err != nil {
				log.Println("unmarshal event err:", err)
				continue
			}

			if event.Type != "order_added" {
				continue
			}

			var order engine.Order
			if err := json.Unmarshal(event.Payload, &order); err != nil {
				log.Println("unmarshal trade err:", err)
				continue
			}

			persistOrder(ctx, db, order)
		}
	}()
}

func StartTradeEventKafkaConsumer(brokers []string, topic string, db *pgxpool.Pool) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "trade_handler",
	})

	go func() {
		ctx := context.Background()
		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Println("kafka read err:", err)
				continue
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
				continue
			}

			var trade engine.Trade
			if err := json.Unmarshal(event.Payload, &trade); err != nil {
				log.Println("unmarshal trade err:", err)
				continue
			}

			if trade.BuyOrderID != "" {
				removeOrderIfFilled(ctx, db, trade.BuyOrderID, trade.Quantity)
			}
			if trade.SellOrderID != "" {
				removeOrderIfFilled(ctx, db, trade.SellOrderID, trade.Quantity)
			}
		}
	}()
}

func persistOrder(ctx context.Context, db *pgxpool.Pool, order engine.Order) {
	_, err := db.Exec(ctx, `INSERT INTO orders (id, symbol, side, price, quantity, remaining, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		order.ID, order.Symbol, order.Side, order.Price, order.Quantity, order.Remaining, time.Now().UTC())
	if err != nil {
		log.Println("persist order err:", err)
	} else {
		log.Printf("Order %s persisted/updated with remaining %f", order.ID, order.Remaining)
	}
}

func removeOrderIfFilled(ctx context.Context, db *pgxpool.Pool, orderID string, filledQty float64) {
	var remaining float64
	err := db.QueryRow(ctx, `SELECT remaining FROM orders WHERE id=$1`, orderID).Scan(&remaining)
	if err != nil {
		return
	}

	remaining -= filledQty
	if remaining <= 0 {
		_, err := db.Exec(ctx, `DELETE FROM orders WHERE id=$1`, orderID)
		if err != nil {
			log.Println("delete order err:", err)
		} else {
			log.Printf("Order %s fully filled and removed", orderID)
		}
	} else {
		_, err := db.Exec(ctx, `UPDATE orders SET remaining=$1 WHERE id=$2`, remaining, orderID)
		if err != nil {
			log.Println("update remaining err:", err)
		}
	}
}
