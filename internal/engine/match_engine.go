package engine

import (
	"context"
	"time"
)

type Engine struct {
	books             map[string]*OrderBook
	orderChannel      chan *Order
	snapshotPublisher EventWriter
	orderPublisher    EventWriter
	tradePublisher    EventWriter
	// processed         int64
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
		books:             make(map[string]*OrderBook),
		orderChannel:      make(chan *Order, 100000),
		snapshotPublisher: eventPublishers[SnapshotTopic],
		orderPublisher:    eventPublishers[OrderTopic],
		tradePublisher:    eventPublishers[TradeTopic],
	}
}

func (engine *Engine) Setup(orderbooks map[string]*OrderBook) {
	engine.books = orderbooks
}

func (engine *Engine) Start(ctx context.Context, isReplay bool) {
	go func() {
		for {
			select {
			case order := <-engine.orderChannel:

				orderbook := engine.GetBook(order.Symbol)
				trades := orderbook.MatchIncoming(order)

				if !isReplay {
					for _, trade := range trades {
						go engine.publishTradeEvent("order_matched", trade)
					}

					if order.Remaining > 0 && order.Price > 0 {
						go engine.publishOrderEvent("order_added", order)
					}
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

func (e *Engine) publishSnapshotEvent(eventType string, payload any) {
	if e.snapshotPublisher == nil {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		books := e.GetBooks()
		for _, book := range books {
			snapshot := book.Snapshot(10)
			e.snapshotPublisher.Publish(eventType, snapshot)
		}

		e.snapshotPublisher.Publish(eventType, payload)
	}
}

// func (e *Engine) Monitor() {
// 	ticker := time.NewTicker(time.Second)
// 	for range ticker.C {
// 		fmt.Println("Emir/saniye:", e.processed)
// 		e.processed = 0
// 	}
// }
