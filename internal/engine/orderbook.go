package engine

import (
	"sync"
)

type OrderBook struct {
    Symbol string
    buys   map[float64]*PriceLevel // price -> level (buy sorted descending)
    buysPrices []float64
    sells  map[float64]*PriceLevel // price -> level (sell sorted ascending)
    sellsPrices []float64
    ordersIndex map[string]*Order   // quick lookup for cancel
    mu sync.RWMutex
}

func NewOrderBook(symbol string) *OrderBook {
    return &OrderBook{
        Symbol: symbol,
        buys: make(map[float64]*PriceLevel),
        sells: make(map[float64]*PriceLevel),
        ordersIndex: make(map[string]*Order),
    }
}
