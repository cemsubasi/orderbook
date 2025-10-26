package ws

import (
	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/gin-gonic/gin"
)

func HandleEventController(r *gin.Engine, e *engine.Engine, h *WsHub) {
	r.GET("/event", func(c *gin.Context) {
		h.HandleWs(c)
	})
}
