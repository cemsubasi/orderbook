package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cemsubasi/orderbook/internal/engine"
	"github.com/rs/xid"
)

type orderCreateRequest struct {
		Symbol string  `json:"symbol" binding:"required"`
		Side   string  `json:"side" binding:"required"`
		Price  float64 `json:"price"` // 0 for market
		Quantity float64 `json:"quantity" binding:"required"`
}

func HandleOrderController(e *engine.Engine) {
	http.HandleFunc("POST /order", func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			responseWriter.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var orderRequest orderCreateRequest
		err := json.NewDecoder(request.Body).Decode(&orderRequest)
		if err != nil {
			responseWriter.WriteHeader(http.StatusBadRequest)
			return
		}

		if orderRequest.Symbol == "" || orderRequest.Side == "" || orderRequest.Quantity <= 0 {
			responseWriter.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(responseWriter, "Missing or invalid fields")
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

