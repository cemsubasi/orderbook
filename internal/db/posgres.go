package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPostgres(pgUser string, pgPass string) *pgxpool.Pool {
    ctx := context.Background()
    dsn := fmt.Sprintf("postgres://%s:%s@localhost:5432/orderbook", pgUser, pgPass)
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        log.Println("pgx pool err:", err)
        return nil
    }

    return pool
}
