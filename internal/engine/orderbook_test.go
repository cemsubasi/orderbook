package engine

import (
	"testing"
)

func TestMatchIncoming_LimitBuyOrderAddedToBookWhenUnfilled(t *testing.T) {
	ob := NewOrderBook("SYM")

	// No sellers, limit buy should be added to buy book
	in := &Order{ID: "o1", Symbol: "SYM", Side: Buy, Price: 50, Remaining: 3}
	trades := ob.MatchIncoming(in)

	if len(trades) != 0 {
		t.Errorf("expected 0 trades, got %d", len(trades))
	}

	// Order should be enqueued into buys and present in ordersIndex
	level, exists := ob.buys[50]
	if !exists {
		t.Fatalf("expected buy price level 50 to exist")
	}
	found := false
	for _, o := range level.Orders {
		if o.ID == "o1" {
			found = true
			if o.Remaining != 3 {
				t.Errorf("expected order remaining 3, got %v", o.Remaining)
			}
		}
	}
	if !found {
		t.Fatalf("expected order o1 in buy level")
	}
}

func TestMatchIncoming_LimitSellOrderAddedToBookWhenUnfilled(t *testing.T) {
	ob := NewOrderBook("SYM")

	// No buyers, limit sell should be added to sell book
	in := &Order{ID: "o1", Symbol: "SYM", Side: Sell, Price: 50, Remaining: 3}
	trades := ob.MatchIncoming(in)

	if len(trades) != 0 {
		t.Errorf("expected 0 trades, got %d", len(trades))
	}

	// Order should be enqueued into sells and present in ordersIndex
	level, exists := ob.sells[50]
	if !exists {
		t.Fatalf("expected sell price level 50 to exist")
	}
	found := false
	for _, o := range level.Orders {
		if o.ID == "o1" {
			found = true
			if o.Remaining != 3 {
				t.Errorf("expected order remaining 3, got %v", o.Remaining)
			}
		}
	}
	if !found {
		t.Fatalf("expected order o1 in buy level")
	}
}

func TestMatchIncoming_BuyMatchesSells(t *testing.T) {
	orderbook := NewOrderBook("SYM")

	// Two sell makers at price 100: m1 qty 1, m2 qty 2
	m1 := &Order{ID: "m1", Symbol: "SYM", Side: Sell, Price: 100, Remaining: 1}
	m2 := &Order{ID: "m2", Symbol: "SYM", Side: Sell, Price: 100, Remaining: 2}

	orderbook.addPriceIfMissing(orderbook.sells, 100, false)
	orderbook.sells[100].Enqueue(m1)
	orderbook.sells[100].Enqueue(m2)

	// incoming buy for 2.5 at price 100 should match m1 fully and m2 partially
	in := &Order{ID: "t1", Symbol: "SYM", Side: Buy, Price: 100, Remaining: 2.5}
	trades := orderbook.MatchIncoming(in)

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(trades))
	}

	// first trade should be 1 (m1), second 1.5 (m2)
	if trades[0].Quantity != 1 {
		t.Errorf("expected first trade qty 1, got %v", trades[0].Quantity)
	}
	if trades[1].Quantity != 1.5 {
		t.Errorf("expected second trade qty 1.5, got %v", trades[1].Quantity)
	}

	// m1 should be removed, m2 should remain with 0.5
	if _, exists := orderbook.sells[100]; !exists {
		t.Fatalf("expected price level 100 to exist")
	}
	// find m2 in level and check remaining
	foundM2 := false
	for _, o := range orderbook.sells[100].Orders {
		if o.ID == "m2" {
			foundM2 = true
			if o.Remaining != 0.5 {
				t.Errorf("expected m2 remaining 0.5, got %v", o.Remaining)
			}
		}
		if o.ID == "m1" {
			t.Errorf("expected m1 removed but found in orders")
		}
	}
	if !foundM2 {
		t.Fatalf("m2 not found in price level after matching")
	}

	if in.Remaining != 0 {
		t.Errorf("expected incoming remaining 0, got %v", in.Remaining)
	}
}

func TestMatchIncoming_SellMatchesBuys(t *testing.T) {
	ob := NewOrderBook("SYM")

	// Two buy makers at price 100: b1 qty 1.5, b2 qty 1
	b1 := &Order{ID: "b1", Symbol: "SYM", Side: Buy, Price: 100, Remaining: 1.5}
	b2 := &Order{ID: "b2", Symbol: "SYM", Side: Buy, Price: 100, Remaining: 1}

	ob.addPriceIfMissing(ob.buys, 100, true)
	ob.buys[100].Enqueue(b1)
	ob.buys[100].Enqueue(b2)

	// incoming sell for 2 at price 100 should match b1 fully (1.5) and b2 partially (0.5)
	in := &Order{ID: "s1", Symbol: "SYM", Side: Sell, Price: 100, Remaining: 2}
	trades := ob.MatchIncoming(in)

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(trades))
	}
	if trades[0].Quantity != 1.5 {
		t.Errorf("expected first trade qty 1.5, got %v", trades[0].Quantity)
	}
	if trades[1].Quantity != 0.5 {
		t.Errorf("expected second trade qty 0.5, got %v", trades[1].Quantity)
	}

	// b1 removed, b2 should remain with 0.5
	foundB2 := false
	for _, o := range ob.buys[100].Orders {
		if o.ID == "b2" {
			foundB2 = true
			if o.Remaining != 0.5 {
				t.Errorf("expected b2 remaining 0.5, got %v", o.Remaining)
			}
		}
		if o.ID == "b1" {
			t.Errorf("expected b1 removed but found in orders")
		}
	}
	if !foundB2 {
		t.Fatalf("b2 not found in buy level after matching")
	}
	if in.Remaining != 0 {
		t.Errorf("expected incoming remaining 0, got %v", in.Remaining)
	}
}
