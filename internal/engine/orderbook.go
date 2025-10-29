package engine

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

type OrderBook struct {
	Symbol      string
	buys        map[float64]*PriceLevel
	buysPrices  []float64
	sells       map[float64]*PriceLevel
	sellsPrices []float64
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		buys:   make(map[float64]*PriceLevel),
		sells:  make(map[float64]*PriceLevel),
	}
}

func (engine *Engine) GetBook(symbol string) *OrderBook {

	book, ok := engine.books[symbol]
	if !ok {
		book = NewOrderBook(symbol)
		engine.books[symbol] = book
	}

	return book
}

func (engine *Engine) GetBooks() map[string]*OrderBook {
	if len(engine.books) == 0 {
		return make(map[string]*OrderBook)
	}

	return engine.books
}

func (orderbook *OrderBook) MatchIncoming(order *Order) []*Trade {
	var trades []*Trade
	remaining := order.Remaining
	if order.Side == Buy {
		for len(orderbook.sellsPrices) > 0 && remaining > 0 {
			bestPrice := orderbook.sellsPrices[0]
			priceLevel := orderbook.sells[bestPrice]
			for len(priceLevel.Orders) > 0 && remaining > 0 && (order.Price == 0 || bestPrice <= order.Price) {
				maker := priceLevel.Peek()
				execQuantity := math.Min(remaining, maker.Remaining)
				trade := &Trade{
					ID:          uuid.New().String(),
					Symbol:      order.Symbol,
					BuyOrderID:  order.ID,
					SellOrderID: maker.ID,
					Price:       maker.Price,
					Quantity:    execQuantity,
					ExecutedAt:  time.Now().UTC(),
				}

				trades = append(trades, trade)
				maker.Remaining -= execQuantity
				remaining -= execQuantity
				if maker.Remaining <= 0 {
					priceLevel.Dequeue()
				}
			}

			if len(priceLevel.Orders) == 0 {
				orderbook.RemovePriceIfEmpty(orderbook.sells, bestPrice, false)
			} else {
				if remaining <= 0 || (order.Price > 0 && bestPrice > order.Price) {
					break
				}
			}
		}
	} else {
		for len(orderbook.buysPrices) > 0 && remaining > 0 {
			bestPrice := orderbook.buysPrices[0]
			priceLevel := orderbook.buys[bestPrice]
			for len(priceLevel.Orders) > 0 && remaining > 0 && (order.Price == 0 || bestPrice >= order.Price) {
				maker := priceLevel.Peek()
				execQuantity := math.Min(remaining, maker.Remaining)
				trade := &Trade{
					ID:          uuid.New().String(),
					Symbol:      order.Symbol,
					BuyOrderID:  maker.ID,
					SellOrderID: order.ID,
					Price:       maker.Price,
					Quantity:    execQuantity,
					ExecutedAt:  time.Now().UTC(),
				}

				trades = append(trades, trade)
				maker.Remaining -= execQuantity
				remaining -= execQuantity
				if maker.Remaining <= 0 {
					priceLevel.Dequeue()
				}
			}

			if len(priceLevel.Orders) == 0 {
				orderbook.RemovePriceIfEmpty(orderbook.buys, bestPrice, true)
			} else {
				if remaining <= 0 || (order.Price > 0 && bestPrice < order.Price) {
					break
				}
			}
		}
	}

	order.Remaining = remaining
	if order.Price > 0 && order.Remaining > 0 {
		if order.Side == Buy {
			orderbook.addPriceIfMissing(orderbook.buys, order.Price, true)
			orderbook.buys[order.Price].Enqueue(order)
		} else {
			orderbook.addPriceIfMissing(orderbook.sells, order.Price, false)
			orderbook.sells[order.Price].Enqueue(order)
		}
	}

	return trades
}

func (orderBook *OrderBook) RemovePriceIfEmpty(priceLevels map[float64]*PriceLevel, price float64, isBuy bool) {
	priceLevel := priceLevels[price]
	if priceLevel != nil && len(priceLevel.Orders) == 0 {
		delete(priceLevels, price)
		if isBuy {
			newPrice := make([]float64, 0, len(orderBook.buysPrices))
			for _, buyPrice := range orderBook.buysPrices {
				if buyPrice != price {
					newPrice = append(newPrice, buyPrice)
				}
			}

			orderBook.buysPrices = newPrice
		} else {
			newPrice := make([]float64, 0, len(orderBook.sellsPrices))
			for _, sellPrice := range orderBook.sellsPrices {
				if sellPrice != price {
					newPrice = append(newPrice, sellPrice)
				}
			}

			orderBook.sellsPrices = newPrice
		}
	}
}

func (orderBook *OrderBook) addPriceIfMissing(priceLevels map[float64]*PriceLevel, price float64, isBuy bool) {
	if _, ok := priceLevels[price]; ok {
		return
	}
	priceLevels[price] = &PriceLevel{Price: price}
	if isBuy {
		orderBook.buysPrices = append(orderBook.buysPrices, price)
		sort.Slice(orderBook.buysPrices, func(i, j int) bool { return orderBook.buysPrices[i] > orderBook.buysPrices[j] })
	} else {
		orderBook.sellsPrices = append(orderBook.sellsPrices, price)
		sort.Float64s(orderBook.sellsPrices)
	}
}

func (orderbook *OrderBook) Snapshot(depth int) (bids []map[string]any, asks []map[string]any) {
	for i, priceLevel := range orderbook.buysPrices {
		if i >= depth {
			break
		}

		volume := 0.0
		for _, order := range orderbook.buys[priceLevel].Orders {
			volume += order.Remaining
		}

		bids = append(bids, map[string]any{"price": priceLevel, "qty": volume})
	}

	for i, price := range orderbook.sellsPrices {
		if i >= depth {
			break
		}

		volume := 0.0
		for _, order := range orderbook.sells[price].Orders {
			volume += order.Remaining
		}

		asks = append(asks, map[string]any{"price": price, "qty": volume})
	}

	return
}

func (ob *OrderBook) AddOrder(order *Order) {
	var levels map[float64]*PriceLevel
	var prices *[]float64

	if order.Side == Buy {
		levels = ob.buys
		prices = &ob.buysPrices
	} else {
		levels = ob.sells
		prices = &ob.sellsPrices
	}

	level, exists := levels[order.Price]
	if !exists {
		level = &PriceLevel{}
		level.Price = order.Price
		levels[order.Price] = level
		*prices = append(*prices, order.Price)

		if order.Side == Buy {
			sort.Slice(*prices, func(i, j int) bool { return (*prices)[i] > (*prices)[j] })
		} else {
			sort.Slice(*prices, func(i, j int) bool { return (*prices)[i] < (*prices)[j] })
		}
	}

	level.Orders = append(level.Orders, order)
}

func SortOrderbooks(orderbooks map[string]*OrderBook) {
	for _, ob := range orderbooks {
		sort.Slice(ob.buysPrices, func(i, j int) bool {
			return ob.buysPrices[i] > ob.buysPrices[j]
		})
		sort.Slice(ob.sellsPrices, func(i, j int) bool {
			return ob.sellsPrices[i] < ob.sellsPrices[j]
		})
	}
}
