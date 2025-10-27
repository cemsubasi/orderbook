package ws

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WsHub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewWsHub() *WsHub {
	return &WsHub{clients: make(map[*websocket.Conn]bool)}
}

func (hub *WsHub) Broadcast(msg []byte) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	for c := range hub.clients {
		_ = c.WriteMessage(websocket.TextMessage, msg)
	}
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func (hub *WsHub) HandleWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	hub.mu.Lock()
	hub.clients[conn] = true
	hub.mu.Unlock()
}
