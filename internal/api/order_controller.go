package api

import (
	"net/http"
	"time"

	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

type orderCreateRequest struct {
		Symbol string  `json:"symbol" binding:"required"`
		Side   string  `json:"side" binding:"required"`
		Price  float64 `json:"price"` // 0 for market
		Quantity float64 `json:"quantity" binding:"required"`
}

func HandleOrderController(r *gin.Engine, e *engine.Engine) {
	r.POST("POST /order", func(c *gin.Context) {
		var orderRequest orderCreateRequest

		if err := c.ShouldBindBodyWithJSON(&orderRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id := xid.New().String()
		order := &engine.Order{
				ID: id,
				Symbol: orderRequest.Symbol,
				Side: engine.Side(orderRequest.Side),
				Price: orderRequest.Price,
				Quantity: orderRequest.Quantity,
				Remaining: orderRequest.Quantity,
				CreatedAt: time.Now().UTC(),
		}

		e.Submit(order)
	})
}

