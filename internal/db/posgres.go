package db

import (
	"context"
	"fmt"
	"log"

	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPostgres(pgUser string, pgPass string, pgHost string, pgPort string, pgDB string) *pgxpool.Pool {
	ctx := context.Background()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pgUser, pgPass, pgHost, pgPort, pgDB)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Println("pgx pool err:", err)
		return nil
	}

	return pool
}

func LoadOrderBooks(pool *pgxpool.Pool) (map[string]*engine.OrderBook, error) {
	rows, err := pool.Query(context.Background(), `
		SELECT id, symbol, side, price, quantity, remaining
		FROM orders
		WHERE remaining > 0
		ORDER BY symbol, 
			CASE WHEN side = 'buy' THEN -price ELSE price END ASC,
			created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("query orders err: %w", err)
	}
	defer rows.Close()

	orderBooks := make(map[string]*engine.OrderBook)

	for rows.Next() {
		var order engine.Order
		if err := rows.Scan(&order.ID, &order.Symbol, &order.Side, &order.Price, &order.Quantity, &order.Remaining); err != nil {
			log.Println("scan order err:", err)
			continue
		}

		ob, exists := orderBooks[order.Symbol]
		if !exists {
			ob = engine.NewOrderBook(order.Symbol)
			orderBooks[order.Symbol] = ob
		}

		ob.AddOrder(&order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	engine.SortOrderbooks(orderBooks)

	return orderBooks, nil
}
