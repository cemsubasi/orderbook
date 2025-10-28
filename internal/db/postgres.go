package db

import (
	"context"
	"fmt"
	"log"

	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPostgres(pgUser string, pgPass string, pgHost string, pgDB string) *pgxpool.Pool {
	ctx := context.Background()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", pgUser, pgPass, pgHost, pgDB)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Println("pgx pool err:", err)
		return nil
	}

	return pool
}

func RetrieveOrderBooks(pool *pgxpool.Pool) (map[string]*engine.OrderBook, error) {
	rows, err := pool.Query(context.Background(), `
       SELECT o.id, o.symbol, o.side, o.price, o.quantity,
       (o.quantity - COALESCE(SUM(
           CASE 
               WHEN o.id = t.buy_order_id THEN t.quantity
               WHEN o.id = t.sell_order_id THEN t.quantity
               ELSE 0
           END
       ),0)) AS remaining
FROM orders o
LEFT JOIN trades t
       ON o.id = t.buy_order_id OR o.id = t.sell_order_id
GROUP BY o.id
HAVING (o.quantity - COALESCE(SUM(
           CASE 
               WHEN o.id = t.buy_order_id THEN t.quantity
               WHEN o.id = t.sell_order_id THEN t.quantity
               ELSE 0
           END
       ),0)) > 0
ORDER BY o.symbol,
         CASE WHEN o.side='buy' THEN -o.price ELSE o.price END,
         o.created_at;
    `)
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
