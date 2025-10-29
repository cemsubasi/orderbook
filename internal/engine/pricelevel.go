package engine

type PriceLevel struct {
	Price  float64
	Orders []*Order
}

func (priceLevel *PriceLevel) Peek() *Order {
	if len(priceLevel.Orders) == 0 {
		return nil
	}
	return priceLevel.Orders[0]
}

func (priceLevel *PriceLevel) Dequeue() *Order {
	if len(priceLevel.Orders) == 0 {
		return nil
	}
	order := priceLevel.Orders[0]
	priceLevel.Orders = priceLevel.Orders[1:]

	return order
}

func (priceLevel *PriceLevel) Enqueue(order *Order) {
	priceLevel.Orders = append(priceLevel.Orders, order)
}
