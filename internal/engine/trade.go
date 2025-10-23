package engine

import "time"

type Trade struct {
    ID string `json:"id"` 
    Symbol string `json:"symbol"`
    BuyOrderID string `json:"buy_order_id"`
    SellOrderID string `json:"sel_order_id"`
    Price float64 `json:"price"`
    Quantity float64 `json:"quantity"`
    ExecutedAt time.Time `json:"executed_at"`
}
