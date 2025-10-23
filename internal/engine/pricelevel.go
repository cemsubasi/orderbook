package engine

type PriceLevel struct {
    Price  float64
    Orders []*Order // FIFO
}

