package handlers

import (
	"database.sql"
	"net/http"
	"errors"

	"github.com/google/uuid"
	"github.com/herodragmon/scalable-ecommerce/services/cart-service/internal/config"
	"github.com/herodragmon/scalable-ecommerce/services/cart-service/internal/response"
)

func RegisterRoutes(mux *http.ServeMux, cfg *config.Config) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "cart-service"})
	})

	mux.HandleFunc("GET /api/cart", func(w http.ResponseWriter, r *http.Request) {
		handlerCartGet(cfg, w, r)
	})
}


func handlerCartGet(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	type CartItemResponse struct {
    ID         uuid.UUID `json:"id"`
    ProductID  uuid.UUID `json:"product_id"`
    Quantity   int32     `json:"quantity"`
    PriceCents int32     `json:"price_cents"`
	}

	type CartResponse struct {
    ID         uuid.UUID          `json:"id,omitempty"`
    Items      []CartItemResponse `json:"items"`
    TotalCents int64              `json:"total_cents"`
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		response.RespondWithError(w, http.StatusUnauthorized, "missing user ID", nil)
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, "invalid user ID", err)
		return
	}
	
	cart, err := cfg.DB.GetCartByUserID(r.Context(), userID)
	if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        response.RespondWithJSON(w, http.StatusOK, CartResponse{
            Items:      []CartItemResponse{},
            TotalCents: 0,
        })
        return
    }
    response.RespondWithError(w, http.StatusInternalServerError, "couldn't get cart", err)
    return
	}

	items, err := cfg.DB.GetCartItems(r.Context(), cart.ID)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "couldn't get cart items", err)
  	return
	}

	var totalCents int64
    itemResponses := make([]CartItemResponse, len(items))
    for i, item := range items {
			totalCents += int64(item.PriceCents) * int64(item.Quantity)
      itemResponses[i] = CartItemResponse{
				ID:         item.ID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
				PriceCents: item.PriceCents,
			}
    }
    
  response.RespondWithJSON(w, http.StatusOK, CartResponse{
		ID:         cart.ID,
		Items:      itemResponses,
		TotalCents: totalCents,
	})
}
