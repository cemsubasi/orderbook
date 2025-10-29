package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/cemsubasi/orderbook/internal/engine"
)

func StartWsSnapshotWorker(hub *WsHub, engine *engine.Engine, ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				snapshot := map[string]map[string]any{}
				books := engine.GetBooks()
				for _, book := range books {
					bids, asks := book.Snapshot(10)
					snapshot[book.Symbol] = map[string]any{
						"bids": bids,
						"asks": asks,
					}
				}

				payload, err := json.Marshal(map[string]any{
					"type":    "snapshot",
					"payload": snapshot,
				})
				if err != nil {
					log.Println("snapshot marshal error:", err)
					continue
				}

				hub.Broadcast(payload)
			}
		}
	}()
}
