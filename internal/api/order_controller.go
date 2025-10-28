package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type orderCreateRequest struct {
	Symbol   string  `json:"symbol" binding:"required"`
	Side     string  `json:"side" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required"`
}

func HandleOrderController(r *gin.Engine, e *engine.Engine) {
	r.POST("/orders", func(c *gin.Context) {
		var orderRequest orderCreateRequest

		if err := c.ShouldBindBodyWithJSON(&orderRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		symbol := strings.ToUpper(strings.TrimSpace(orderRequest.Symbol))
		side := strings.ToLower(strings.TrimSpace(orderRequest.Side))
		price := orderRequest.Price
		qty := orderRequest.Quantity

		if symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
			return
		}

		if side != "buy" && side != "sell" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "side must be 'buy' or 'sell'"})
			return
		}

		if price <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price must be greater than zero"})
			return
		}

		if qty <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be greater than zero"})
			return
		}

		if price > 1e9 || qty > 1e9 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price/quantity too large"})
			return
		}

		id := uuid.New().String()
		order := &engine.Order{
			ID:        id,
			Symbol:    orderRequest.Symbol,
			Side:      engine.Side(orderRequest.Side),
			Price:     orderRequest.Price,
			Quantity:  orderRequest.Quantity,
			Remaining: orderRequest.Quantity,
			CreatedAt: time.Now().UTC(),
		}

		e.Submit(order)
		c.JSON(http.StatusAccepted, gin.H{"orderId": id})
	})

	r.GET("/orderbook/:symbol", func(c *gin.Context) {
		symbol := c.Param("symbol")
		depthQ := c.Query("depth")
		depth := 10
		if depthQ != "" {
			fmt.Sscanf(depthQ, "%d", &depth)
		}

		book := e.GetBook(symbol)
		bids, asks := book.Snapshot(depth)
		c.JSON(http.StatusOK, gin.H{"symbol": symbol, "bids": bids, "asks": asks})
	})

	r.GET("/orderbook", func(c *gin.Context) {
		depthQ := c.Query("depth")
		depth := 10
		if depthQ != "" {
			fmt.Sscanf(depthQ, "%d", &depth)
		}

		books := e.GetBooks()
		result := make(map[string]gin.H)
		for sym, book := range books {
			bids, asks := book.Snapshot(depth)
			result[sym] = gin.H{"bids": bids, "asks": asks}
		}

		c.JSON(http.StatusOK, result)
	})
}
