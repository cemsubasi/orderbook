package engine

import (
	"context"
	"log"
)

type Engine struct {
	books          map[string]*OrderBook
	orderChannel   chan *Order
	eventPublisher EventWriter
}

type EventWriter interface {
	Publish(eventType string, payload any) error
}

func NewEngine(eventPublisher EventWriter) *Engine {
	return &Engine{
		books:          make(map[string]*OrderBook),
		orderChannel:   make(chan *Order, 1024),
		eventPublisher: eventPublisher,
	}
}

func (engine *Engine) Setup(orderbooks map[string]*OrderBook) {
	engine.books = orderbooks
}

func (engine *Engine) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case order := <-engine.orderChannel:
				log.Printf("Incoming OrderID is: %s, OrderSymbol is: %s", order.ID, order.Symbol)

				orderbook := engine.GetBook(order.Symbol)
				trades := orderbook.MatchIncoming(order)
				for _, trade := range trades {
					engine.publishEvent("order_matched", trade)
				}

				if order.Remaining > 0 && order.Price > 0 {
					engine.publishEvent("order_added", order)
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}

func (engine *Engine) Submit(order *Order) {
	engine.orderChannel <- order
}

func (e *Engine) publishEvent(eventType string, payload any) {
	if e.eventPublisher == nil {
		return
	}
	_ = e.eventPublisher.Publish(eventType, payload)
}
