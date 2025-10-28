package engine

import (
	"context"
	"fmt"
	"time"
)

type Engine struct {
	books          map[string]*OrderBook
	orderChannel   chan *Order
	eventPublisher EventWriter
	processed      int64
}

type EventWriter interface {
	Publish(eventType string, payload any) error
}

func NewEngine(eventPublisher EventWriter) *Engine {
	return &Engine{
		books:          make(map[string]*OrderBook),
		orderChannel:   make(chan *Order, 100000),
		eventPublisher: eventPublisher,
		// processed:      0,
	}
}

func (engine *Engine) Setup(orderbooks map[string]*OrderBook) {
	engine.books = orderbooks
}

type PartialMatchedOrder struct {
	Order  *Order
	Trades []*Trade
}

func (engine *Engine) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case order := <-engine.orderChannel:

				orderbook := engine.GetBook(order.Symbol)
				trades := orderbook.MatchIncoming(order)

				for _, trade := range trades {
					go engine.publishEvent("order_matched", trade)
				}

				if order.Remaining > 0 && order.Price > 0 {
					go engine.publishEvent("order_added", order)
				}

				// atomic.AddInt64(&engine.processed, 1)

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
	e.eventPublisher.Publish(eventType, payload)
}

func (e *Engine) Monitor() {
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		fmt.Println("Emir/saniye:", e.processed)
		e.processed = 0
	}
}
