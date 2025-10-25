package engine

import (
	"context"
	"log"
)

type Engine struct {
	books        map[string]*OrderBook
	orderChannel chan *Order
}

func NewEngine() *Engine {
	return &Engine{
		books:        make(map[string]*OrderBook),
		orderChannel: make(chan *Order, 1024),
	}
}

func (engine *Engine) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case order := <-engine.orderChannel:
				log.Printf("OrderID is: %s, OrderSymbol is: %s", order.ID, order.Symbol)
				orderbook := engine.GetBook(order.Symbol)
				orderbook.MatchIncoming(order)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (engine *Engine) Submit(order *Order) {
	engine.orderChannel <- order
}
