package engine

import (
	"context"
)

type Engine struct {
	books          map[string]*OrderBook
	orderChannel   chan *Order
	orderPublisher EventWriter
	tradePublisher EventWriter
}

type EventWriter interface {
	Publish(eventType string, payload any) error
}

const (
	SnapshotTopic = "orderbook_snapshot"
	OrderTopic    = "order_events"
	TradeTopic    = "trade_events"
)

func NewEngine(eventPublishers map[string]EventWriter) *Engine {
	return &Engine{
		books:          make(map[string]*OrderBook),
		orderChannel:   make(chan *Order, 100000),
		orderPublisher: eventPublishers[OrderTopic],
		tradePublisher: eventPublishers[TradeTopic],
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

				orderbook := engine.GetBook(order.Symbol)
				trades := orderbook.MatchIncoming(order)

				if len(trades) > 0 {
					go engine.publishTradeEvent("order_matched", trades)
				}

				if order.Remaining > 0 && order.Price > 0 {
					go engine.publishOrderEvent("order_added", order)
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

func (e *Engine) publishOrderEvent(eventType string, payload any) {
	if e.orderPublisher == nil {
		return
	}
	e.orderPublisher.Publish(eventType, payload)
}

func (e *Engine) publishTradeEvent(eventType string, payload any) {
	if e.tradePublisher == nil {
		return
	}
	e.tradePublisher.Publish(eventType, payload)
}
