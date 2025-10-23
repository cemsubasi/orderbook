package engine

type Engine struct {
    books map[string]*OrderBook
    orderChannel chan *Order
}

func NewEngine() *Engine {
    return &Engine{
        books: make(map[string]*OrderBook),
        orderChannel: make(chan *Order, 1024),
    }
}

