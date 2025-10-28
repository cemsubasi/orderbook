package event

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/cemsubasi/orderbook/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func StartKafkaWsConsumer(hub *ws.WsHub, brokers []string, topic string) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		for {
			log.Println("Creating reader...")
			r := kafka.NewReader(kafka.ReaderConfig{
				Brokers: brokers,
				Topic:   topic,
				GroupID: "ws_broadcast",
			})
			defer r.Close()
			log.Println("Kafka ws consumer connected")

			for {
				m, err := r.ReadMessage(ctx)
				if err != nil {
					log.Println("kafka read err:", err)
					log.Println("Reconnecting to Kafka in 1 second...")
					_ = r.Close()
					time.Sleep(1 * time.Second)
					break
				}

				hub.Broadcast(m.Value)
			}
		}
	}()
}

func StartKafkaOrderConsumer(brokers []string, topic string, db *pgxpool.Pool) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for {
			log.Println("Creating reader...")
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers: brokers,
				Topic:   topic,
				GroupID: "order_handler",
			})
			defer reader.Close()
			log.Println("Kafka order consumer connected")

			for {
				m, err := reader.ReadMessage(ctx)
				if err != nil {
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

				if event.Type != "order_added" {
					return
				}

				var order *engine.Order
				if err := json.Unmarshal(event.Payload, &order); err != nil {
					log.Println("unmarshal trade err:", err)
					continue
				}

				persistOrder(ctx, db, order)
			}
		}
	}()
}

func StartKafkaTradeConsumer(brokers []string, topic string, db *pgxpool.Pool) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		for {
			log.Println("Creating reader...")
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers: brokers,
				Topic:   topic,
				GroupID: "trade_handler",
			})

			defer reader.Close()
			log.Println("Kafka trade consumer connected")

			for {
				m, err := reader.ReadMessage(ctx)
				if err != nil {
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
					continue
				}

				var trade engine.Trade
				if err := json.Unmarshal(event.Payload, &trade); err != nil {
					log.Println("unmarshal trade err:", err)
					continue
				}

				persistTrade(ctx, db, &trade)
			}
		}
	}()
}

func persistOrder(ctx context.Context, db *pgxpool.Pool, order *engine.Order) {
	_, err := db.Exec(ctx, `INSERT INTO orders (id, symbol, side, price, quantity, remaining, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		order.ID, order.Symbol, order.Side, order.Price, order.Quantity, order.Remaining, time.Now().UTC())
	if err != nil {
		log.Println("persist order err:", err)
	} else {
		// log.Printf("Order %s persisted/updated with remaining %f", order.ID, order.Remaining)
	}
}

// func removeOrderIfFilled(ctx context.Context, db *pgxpool.Pool, orderID string, filledQty float64) {
// 	var remaining float64
// 	err := db.QueryRow(ctx, `SELECT remaining FROM orders WHERE id=$1`, orderID).Scan(&remaining)
// 	if err != nil {
// 		return
// 	}

// 	remaining -= filledQty
// 	if remaining <= 0 {
// 		_, err := db.Exec(ctx, `DELETE FROM orders WHERE id=$1`, orderID)
// 		if err != nil {
// 			log.Println("delete order err:", err)
// 		} else {
// 			// log.Printf("Order %s fully filled and removed", orderID)
// 		}
// 	} else {
// 		_, err := db.Exec(ctx, `UPDATE orders SET remaining=$1 WHERE id=$2`, remaining, orderID)
// 		if err != nil {
// 			log.Println("update remaining err:", err)
// 		}
// 	}
// }

func persistTrade(ctx context.Context, db *pgxpool.Pool, trade *engine.Trade) {
	_, err := db.Exec(ctx, `INSERT INTO trades (id, symbol, buy_order_id, sell_order_id, price, quantity, executed_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		trade.ID, trade.Symbol, trade.BuyOrderID, trade.SellOrderID, trade.Price, trade.Quantity, trade.ExecutedAt)
	if err != nil {
		log.Println("persist trade err:", err)
	} else {
		// log.Printf("Trade %s persisted", trade.ID)
	}
}
