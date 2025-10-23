package engine

import (
    "time"
)

type Order struct {
    ID        string    `json:"id"`
    Symbol    string    `json:"symbol"`
    Side      Side      `json:"side"`
    Price     float64   `json:"price"` // price == 0 => market
    Quantity  float64   `json:"quantity"`
    Remaining float64   `json:"remaining"`
    CreatedAt time.Time `json:"created_at"`
}
