package engine

type Engine struct {
    books map[string]*OrderBook
    inCh chan *Order
}

func NewEngine() *Engine {
    return &Engine{
        books: make(map[string]*OrderBook),
        inCh: make(chan *Order, 1024),
    }
}

